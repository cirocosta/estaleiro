package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cirocosta/estaleiro/config"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/pkg/errors"

	dockerfile "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	gw "github.com/moby/buildkit/frontend/gateway/client"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	utilsImage = "docker.io/library/alpine:latest@sha256:1072e499f3f655a032e88542330cf75b02e7bdf673278f701d7ba61629ee3ebe"
)

// TODO receive a context so that image resolution is not unbound
//
func ToLLB(ctx context.Context, cfg *config.Config) (state llb.State, img ocispec.Image, bom Bom, err error) {
	// TODO consider tag provided
	//
	canonicalName, err := resolveImage(ctx, cfg.Image.BaseImage.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to resolve digest for %s when preparing llb",
			cfg.Image.BaseImage.Name)
		return
	}

	bom.Version = "v0.0.1"
	bom.GeneratedAt = time.Now()
	bom.BaseImage = BaseImage{
		Name:   canonicalName.Name(),
		Digest: canonicalName.Digest().String(),
	}

	state = llb.Image(canonicalName.String())
	state = installPackages(state, cfg.Image.Apt)

	tarballStateMap := map[string]llb.State{}
	for _, tarball := range cfg.Tarballs {
		src := llb.Local("context")
		dest := llb.Scratch()

		// how to access the file there?
		tarballStateMap[tarball.Name] = copy(src, tarball.Name, dest, "/dest")
	}

	for _, file := range cfg.Image.Files {
		switch {
		case file.FromStep != nil:
			state, bom, err = copyFilesFromSteps(state, cfg, bom, file)
		case file.FromTarball != nil:
			tarballSourceState, found := tarballStateMap[file.FromTarball.TarballName]
			if !found {
				// TODO improve this
				err = errors.Errorf("not found")
				return
			}

			state, bom, err = copyFilesFromTarball(state, cfg, bom, file, tarballSourceState)
		}

		if err != nil {
			return
		}

	}

	img = prepareImage(cfg.Image)

	return
}

// Copies a file (`file`) from a given state that already has the tarball
// unpacked (`tarballSource`) into the final state (`finalState`)
//
func copyFilesFromTarball(
	finalState llb.State, cfg *config.Config, bom Bom, file config.File, tarballSource llb.State,
) (
	newState llb.State, newBom Bom, err error,
) {
	newState = copy(
		tarballSource,
		"/dest/"+file.FromTarball.Path,
		finalState,
		file.Destination,
	)

	// func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {

	return
}

// prepareImage populates the final definition of the OCI Image spec object
// that can be used to influence the runtime that runs the container image that
// we generate.
//
func prepareImage(image config.Image) ocispec.Image {
	env := []string{}
	for k, v := range image.Env {
		env = append(env, k+"="+v)
	}

	vols := map[string]struct{}{}
	for _, vol := range image.Volumes {
		vols[vol] = struct{}{}
	}

	return ocispec.Image{
		Architecture: "amd64",
		OS:           "linux",
		Config: ocispec.ImageConfig{
			Env:        env,
			Volumes:    vols,
			StopSignal: image.StopSignal,
			Entrypoint: image.Entrypoint,
			Cmd:        image.Cmd,
		},
	}
}

func copyFilesFromSteps(
	state llb.State, cfg *config.Config, bom Bom, file config.File,
) (
	newState llb.State, newBom Bom, err error,
) {
	configStep := getStepFromConfig(cfg, file.FromStep.StepName)
	if configStep == nil {
		err = errors.Errorf("referenced step %s not declared",
			file.FromStep.StepName)
		return
	}

	// the file has a `path`, thus, make sure that the `path` matches a
	// `source_file` in the `step` definition

	var (
		fileFoundInStep = false
		bomFile         = File{Path: file.Destination}
	)

	for _, sourceFile := range configStep.SourceFiles {
		if file.FromStep.Path != sourceFile.Location {
			continue
		}

		fileFoundInStep = true

		bomFile.Source = Source{
			Type:    sourceFile.VCS.Type,
			Version: sourceFile.VCS.Ref,
			Uri:     sourceFile.VCS.Repository,
		}

		bom.Files = append(bom.Files, bomFile)
	}

	if !fileFoundInStep {
		err = errors.Errorf("file %s not declared in step %s",
			file.FromStep.Path, configStep.Name)
		return
	}

	var step llb.State
	step, err = addImageBuildStep(configStep)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to add step to image building process")
		return
	}

	newState = copy(step, file.FromStep.Path, state, file.Destination)
	newBom = bom
	return
}

func getStepFromConfig(cfg *config.Config, name string) *config.Step {
	for _, step := range cfg.Steps {
		if step.Name == name {
			return &step
		}
	}

	return nil
}

// copy copies files between 2 states
//
func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	return dest.File(llb.Copy(src, srcPath, destPath, &llb.CopyInfo{
		AttemptUnpack:  true,
		CreateDestPath: true,
	}))

}

func addImageBuildStep(step *config.Step) (state llb.State, err error) {
	var (
		stepState         *llb.State
		dockerfileContent []byte
	)

	dockerfileContent, err = readFile(step.Dockerfile)
	if err != nil {
		err = errors.Wrapf(err,
			"failed reading dockerfile %s", step.Dockerfile)
		return
	}

	stepState, _, err = dockerfile.Dockerfile2LLB(
		context.TODO(), dockerfileContent, dockerfile.ConvertOpt{
			Target: step.Target,
			BuildPlatforms: []specs.Platform{
				{
					Architecture: "amd64",
					OS:           "linux",
				},
			},
		})
	if err != nil {
		err = errors.Wrapf(err,
			"failed to convert dockerfile to llb")
		return
	}

	state = *stepState
	return
}

// TODO - keep track of these extra utilities that we're installing
//        - could, perhaps, just be providing a `bom` that gets mutated?
//
func installPackages(base llb.State, apts []config.Apt) llb.State {
	// adding two here already
	base = base.Run(shf("apt update && apt install -y apt-transport-https ca-certificates gnupg-agent")).Root()

	// TODO - have all of this hapenning in a single step

	for _, apt := range apts {
		for _, repo := range apt.Repositories {
			base = base.Run(shf("echo \"%s\" >> /etc/apt/sources.list", repo.Uri)).Root()

			if repo.Source != "" {
				base = base.Run(shf("echo \"%s\" >> /etc/apt/sources.list", repo.Source)).Root()
			}
		}

		for _, key := range apt.Keys {
			base = aptAddKey(base, key.Uri)
		}

		if len(apt.Packages) != 0 {
			pkgInstall := "apt update && apt install --no-install-recommends --no-install-suggests -y"

			for _, pkg := range apt.Packages {
				pkgInstall = pkgInstall + " " + pkg.String()
			}

			base = base.Run(sh(pkgInstall)).Root()
			base = base.Run(sh("rm -rf /var/lib/apt/lists/*")).Root()
		}

	}

	return base
}

// TODO get rid of this and leverage the `HTTP` source instead.
//
func curl() llb.State {
	return llb.Image(utilsImage).
		Run(llb.Shlex("apk add --no-cache curl")).Root()
}

func aptAddKey(dst llb.State, url string) llb.State {
	downloadSt := curl().Run(llb.Shlexf("curl -Lo /key.gpg %s", url)).Root()

	dst = copy(downloadSt, "/key.gpg", dst, "/key.gpg")

	return dst.
		Run(sh("apt-key add /key.gpg && rm /key.gpg")).
		Root()
}

func sh(cmd string) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", cmd})
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}

func readFile(filename string) (content []byte, err error) {
	var file *os.File

	file, err = os.Open(filename)
	if err != nil {
		return
	}

	defer file.Close()

	content, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return
}

func resolveImage(ctx context.Context, baseName string) (canonicalName reference.Canonical, err error) {
	var (
		metaResolver = imagemetaresolver.Default()
		ref          reference.Named
		d            digest.Digest
	)

	ref, err = reference.ParseNormalizedNamed(baseName)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse stage name %q", baseName)
		return
	}

	ref = reference.TagNameOnly(ref)
	finalName := ref.String()

	d, _, err = metaResolver.ResolveImageConfig(ctx, finalName, gw.ResolveImageConfigOpt{
		Platform:    &linuxAMD64,
		ResolveMode: llb.ResolveModeDefault.String(),
		LogName:     "resolving",
	})
	if err != nil {
		err = errors.Wrapf(err,
			"couldn't resolve image for %s", finalName)
		return
	}

	canonicalName, err = reference.WithDigest(ref, d)
	if err != nil {
		err = errors.Wrapf(err,
			"couldn't retrieve canonical name")
		return
	}

	return
}
