package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cirocosta/estaleiro/bom"
	"github.com/cirocosta/estaleiro/config"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/moby/buildkit/solver/pb"
	"github.com/pkg/errors"

	dockerfile "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	gw "github.com/moby/buildkit/frontend/gateway/client"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	imageName                  = "cirocosta/estaleiro@sha256:f27dd2a0011116a05346f966c79699a0bb10ff197240af3d90efd11543dfa43a"
	initialPackagesBomFilepath = "/initial-packages-bom.yml"
	finalPackagesBomFilepath   = "/final-packages-bom.yml"
)

func generatePackagesBom(base llb.State, destFilename string) llb.State {
	return base.Run(
		llb.Args([]string{
			"/usr/local/bin/estaleiro",
			"collect",
			"--input=/var/lib/dpkg/status",
			"--output=" + destFilename,
		}), llb.AddMount(
			"/usr/local/bin/estaleiro",
			llb.Image(imageName),
			llb.SourcePath("/usr/local/bin/estaleiro"),
		)).Root()
}

// TODO receive a context so that image resolution is not unbound
//
func ToLLB(ctx context.Context, cfg *config.Config) (state llb.State, img ocispec.Image, materials bom.Bom, err error) {
	// TODO consider tag provided
	//
	canonicalName, err := resolveImage(ctx, cfg.Image.BaseImage.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to resolve digest for %s when preparing llb",
			cfg.Image.BaseImage.Name)
		return
	}

	materials.Version = "v0.0.1"
	materials.GeneratedAt = time.Now()
	materials.BaseImage = bom.BaseImage{
		Name:   canonicalName.Name(),
		Digest: canonicalName.Digest().String(),
	}

	state = llb.Image(canonicalName.String())
	state = generatePackagesBom(state, initialPackagesBomFilepath)
	state = installPackages(state, cfg.Image.Apt)

	// create states for the tarballs where they have their contents
	// already extracted so that we can consume the files later
	//
	tarballStateMap := map[string]llb.State{}
	for _, tarball := range cfg.Tarballs {
		src := llb.Local("context")
		dest := llb.Scratch()

		// what if we did the `untar`ing? :thinking:

		// how to access the file there?
		tarballStateMap[tarball.Name] = copy(src, tarball.Name, dest, "/dest")
	}

	for _, file := range cfg.Image.Files {
		switch {
		case file.FromStep != nil:
			state, materials, err = copyFilesFromSteps(state, cfg, materials, file)
		case file.FromTarball != nil:
			tarballSourceState, found := tarballStateMap[file.FromTarball.TarballName]
			if !found {
				// TODO improve this
				err = errors.Errorf("not found")
				return
			}

			state, materials, err = copyFilesFromTarball(state, cfg, materials, file, tarballSourceState)
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
	finalState llb.State, cfg *config.Config, materials bom.Bom, file config.File, tarballSource llb.State,
) (
	newState llb.State, newBom bom.Bom, err error,
) {
	newState = copy(
		tarballSource,
		"/dest/"+file.FromTarball.Path,
		finalState,
		file.Destination,
	)

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
	state llb.State, cfg *config.Config, materials bom.Bom, file config.File,
) (
	newState llb.State, newBom bom.Bom, err error,
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
		bomFile         = bom.File{Path: file.Destination}
	)

	for _, sourceFile := range configStep.SourceFiles {
		if file.FromStep.Path != sourceFile.Location {
			continue
		}

		fileFoundInStep = true

		bomFile.Source = bom.Source{
			Type:    sourceFile.VCS.Type,
			Version: sourceFile.VCS.Ref,
			Uri:     sourceFile.VCS.Repository,
		}

		materials.ChangeSet.Files = append(materials.ChangeSet.Files, bomFile)
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
	newBom = materials
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

	caps := pb.Caps.CapSet(pb.Caps.All())

	stepState, _, err = dockerfile.Dockerfile2LLB(
		context.TODO(), dockerfileContent, dockerfile.ConvertOpt{
			Target:  step.Target,
			LLBCaps: &caps,
			MetaResolver: imagemetaresolver.New(
				imagemetaresolver.WithDefaultPlatform(&linuxAMD64),
			),
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
	// TODO - have all of this done through the binary
	for _, apt := range apts {
		// TODO - bring repositories back
		//      - these could be added through `estaleiro apt`
		//
		// for _, repo := range apt.Repositories {
		// 	base = base.Run(shf("echo \"%s\" >> /etc/apt/sources.list", repo.Uri)).Root()
		// 	if repo.Source != "" {
		// 		base = base.Run(
		// 			shf("echo \"%s\" >> /etc/apt/sources.list", repo.Source),
		// 		).Root()
		// 	}
		// }

		// TODO - bring keys back
		//
		// for _, key := range apt.Keys {
		// 	base = aptAddKey(base, key.Uri)
		// }

		if len(apt.Packages) == 0 {
			return base
		}

		pkgs := make([]string, len(apt.Packages))
		for idx, pkg := range apt.Packages {
			pkgs[idx] = pkg.String()
		}

		base = base.Run(
			llb.Args(append([]string{
				"/usr/local/bin/estaleiro",
				"apt",
				"--output=" + finalPackagesBomFilepath,
			}, pkgs...)),
			llb.AddMount(
				"/usr/local/bin/estaleiro",
				llb.Image(imageName),
				llb.SourcePath("/usr/local/bin/estaleiro"),
			),
		).Root()
	}

	return base
}

func aptAddKey(dst llb.State, url string) llb.State {
	downloadSt := llb.HTTP(url, llb.Filename("key.gpg"))

	dst = copy(downloadSt, "key.gpg", dst, "/key.gpg")

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
