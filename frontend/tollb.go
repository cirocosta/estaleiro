package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	imageName = "cirocosta/estaleiro@sha256:a2dc7d2c4bde47afa6f3ed312f7f791253f5db9bda2154d0288152829b9546ab"
)

func estaleiroSourceMount() llb.RunOption {
	// TODO base this in `debug` or not

	// return llb.AddMount(
	// 	"/usr/local/bin/estaleiro",
	// 	llb.Image(imageName),
	// 	llb.SourcePath("/usr/local/bin/estaleiro"),
	// )

	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Local("bin"),
		llb.SourcePath("estaleiro"),
	)
}

// gathers information from `os-release` so that in the final bill of materials
// we're able to tell the OS information about the base image.
//
func generateOsReleaseBom(base llb.State, bomState llb.State) llb.State {
	return base.Run(
		llb.Args([]string{
			"/usr/local/bin/estaleiro",
			"base",
			"--output=/bom/base.yml",
		}),
		estaleiroSourceMount(),
	).AddMount("/bom", bomState)
}

// unarchives a given set of files from a source tarball at `srcPath` from a
// `src` state to a `destPath` of a `dest` state.
//
func unarchive(
	bom, src llb.State, srcPath string, dest llb.State, destPath string, files []string,
) (
	llb.State, llb.State,
) {
	var (
		image = llb.Scratch()
		args  = []string{
			"/usr/local/bin/estaleiro",
			"extract",
			`--tarball=` + path.Join("/src", srcPath),
			`--destination=/dest`,
			`--output=/bom/unarchive.yml`,
		}
	)

	for _, file := range files {
		args = append(args, "--file="+file)
	}

	cp := image.Run(llb.Args(args), estaleiroSourceMount())
	cp.AddMount("/src", src, llb.Readonly)

	return cp.AddMount("/dest", dest), cp.AddMount("/bom", bom)
}

// installs a list of packages into `base`, providing a bill of materials at
// `bomState`.
//
func installPackages(base llb.State, bom llb.State, apts []config.Apt) (llb.State, llb.State) {
	allPackages := []string{}

	for _, apt := range apts {
		if len(apt.Packages) == 0 {
			continue
		}

		pkgs := make([]string, len(apt.Packages))
		for idx, pkg := range apt.Packages {
			pkgs[idx] = pkg.String()
		}

		allPackages = append(allPackages, pkgs...)

	}

	if len(allPackages) == 0 {
		return base, bom
	}

	run := base.Run(
		llb.Args(append([]string{
			"/usr/local/bin/estaleiro",
			"apt",
			"--output=/bom/final-packages.yml",
		}, allPackages...)),
		estaleiroSourceMount(),
	)

	return run.Root(), run.AddMount("/bom", bom)
}

// gathers the package listing from a given state, saving the bill of materials
// in the filesystem at `destFilename`.
//
func generatePackagesBom(base llb.State, bomState llb.State) llb.State {
	return base.Run(
		llb.Args([]string{
			"/usr/local/bin/estaleiro",
			"collect",
			"--input=/var/lib/dpkg/status",
			`--output=/bom/initial-packages.yml`,
		}),
		estaleiroSourceMount(),
	).AddMount("/bom", bomState)
}

func tarballFiles(fs, bom llb.State, cfg *config.Config) (newFs llb.State, newBom llb.State) {
	newFs, newBom = fs, bom
	extractionsMap := map[string][]string{}

	for _, file := range cfg.Image.Files {
		if file.FromTarball == nil {
			continue
		}

		filesToExtract, _ := extractionsMap[file.FromTarball.TarballName]
		extractionsMap[file.FromTarball.TarballName] = append(filesToExtract, file.FromTarball.Path)
	}

	tarballStateMap := make(map[string]llb.State, len(extractionsMap))
	for tarball, files := range extractionsMap {
		var (
			src  = llb.Local("context")
			dest = llb.Scratch()
		)

		tarballStateMap[tarball], newBom = unarchive(
			newBom, src, tarball, dest, "/dest", files,
		)
	}

	for _, file := range cfg.Image.Files {
		if file.FromTarball == nil {
			continue
		}

		tarballSourceState, found := tarballStateMap[file.FromTarball.TarballName]
		if !found {
			panic(errors.Errorf("not found"))
		}

		newFs = copyFilesFromTarball(newFs, file, tarballSourceState)
	}

	return
}

// TODO emit new bom
//
func runAndCopyFromSteps(fs llb.State, cfg *config.Config) (newFs llb.State, err error) {
	newFs = fs

	for _, file := range cfg.Image.Files {
		if file.FromStep == nil {
			continue
		}

		newFs, err = copyFilesFromSteps(newFs, cfg, file)
		if err != nil {
			// TODO better error
			return
		}
	}

	return
}

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

	bomState := llb.Scratch()

	materials.Version = "v0.0.1"
	materials.GeneratedAt = time.Now()

	materials.BaseImage = bom.BaseImage{
		Name:   canonicalName.Name(),
		Digest: canonicalName.Digest().String(),
	}

	state = llb.Image(canonicalName.String())

	bomState = generatePackagesBom(state, bomState)
	bomState = generateOsReleaseBom(state, bomState)
	state, bomState = installPackages(state, bomState, cfg.Image.Apt)
	state, bomState = tarballFiles(state, bomState, cfg)

	state, err = runAndCopyFromSteps(state, cfg)
	if err != nil {
		// TODO improve
		return
	}

	state = copy(bomState, "*.yml", state, "/bom")

	img = prepareImage(cfg.Image)

	return
}

// Copies a file (`file`) from a given state that already has the tarball
// unpacked (`tarballSource`) into the final state (`finalState`)
//
func copyFilesFromTarball(
	finalState llb.State, file config.File, tarballSource llb.State,
) llb.State {
	return copy(
		tarballSource,
		file.FromTarball.Path,
		finalState,
		file.Destination,
	)
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
	state llb.State, cfg *config.Config, file config.File,
) (
	newState llb.State, err error,
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
		AllowWildcard:  true,
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
