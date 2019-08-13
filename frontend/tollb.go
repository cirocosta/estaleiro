package frontend

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"github.com/cirocosta/estaleiro/config"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/moby/buildkit/solver/pb"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	dockerfile "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	gw "github.com/moby/buildkit/frontend/gateway/client"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func ToLLB(
	ctx context.Context, cfg *config.Config, dockerfileMapping map[string][]byte,
) (
	fs llb.State, img ocispec.Image, err error,
) {
	fs = llb.Scratch()
	bomState := llb.Scratch()
	metadata := bomfs.Meta{Image: "scratch"}

	if cfg.Image.BaseImage.Name != "scratch" {
		var canonicalName reference.Canonical

		canonicalName, err = resolveImage(ctx, cfg.Image.BaseImage.Name)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to resolve digest for %s when preparing llb",
				cfg.Image.BaseImage.Name)
			return
		}

		fs = llb.Image(canonicalName.String())
		metadata.Image = canonicalName.String()
	}

	bomState = generateMetaBom(bomState, metadata)

	if cfg.Image.BaseImage.Name != "scratch" {

		bomState = generatePackagesBom(fs, bomState)
		bomState = generateOsReleaseBom(fs, bomState)

		if len(cfg.Image.Apt) > 0 {
			fs, bomState = packages(fs, bomState, cfg.Image.Apt)

		}
	}

	fs, bomState, err = tarballFiles(fs, bomState, cfg)
	if err != nil {
		return
	}

	fs, bomState, err = runAndCopyFromSteps(fs, bomState, cfg, dockerfileMapping)
	if err != nil {
		return
	}

	bomState = mergeBom(bomState)

	fs = copy(bomState, "merged.yml", fs, "/bom/merged.yml")
	img = prepareImage(cfg.Image)

	return
}

func allKeys(apts []config.Apt) (res []string) {
	for _, apt := range apts {
		if len(apt.Keys) == 0 {
			continue
		}

		arr := make([]string, len(apt.Keys))
		for idx, val := range apt.Keys {
			arr[idx] = val.Uri
		}

		res = append(res, arr...)
	}

	return
}

func allPackages(apts []config.Apt) (res []string) {
	for _, apt := range apts {
		if len(apt.Packages) == 0 {
			continue
		}

		arr := make([]string, len(apt.Packages))
		for idx, val := range apt.Packages {
			arr[idx] = val.String()
		}

		res = append(res, arr...)
	}

	return
}

func allRepositories(apts []config.Apt) (res []string) {
	for _, apt := range apts {
		if len(apt.Repositories) == 0 {
			continue
		}

		arr := make([]string, len(apt.Repositories))
		for idx, val := range apt.Repositories {
			arr[idx] = val
		}

		res = append(res, arr...)
	}

	return
}

func prefixSlice(slice []string, prefix string) (res []string) {
	res = make([]string, len(slice))

	for idx, item := range slice {
		res[idx] = prefix + item
	}

	return
}

func mergeBom(bomState llb.State) llb.State {
	return bomState.Run(
		llb.Args([]string{
			"/usr/local/bin/estaleiro",
			"merge",
			"--directory=.",
			"--output=/merged.yml",
		}),
		estaleiroSourceMount(),
	).Root()
}

func generateMetaBom(bomState llb.State, meta bomfs.Meta) llb.State {
	res, err := yaml.Marshal(bomfs.NewMetaV1(meta))
	if err != nil {
		panic(err)
	}

	return bomState.File(llb.Mkfile("/meta.yml", 0755, res))
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

// packages retrieves and installs the desired packages.
//
// 1. retrieves repositories
// 2. retrieves packages
//
//
func packages(fs, bomState llb.State, apts []config.Apt) (newFs, newBom llb.State) {
	newFs, newBom = fs, bomState

	var (
		keys         = prefixSlice(allKeys(apts), "-k=")
		packages     = prefixSlice(allPackages(apts), "-p=")
		repositories = prefixSlice(allRepositories(apts), "-r=")
	)

	repositoriesState := fs.Run(
		llb.Args(append([]string{
			"/usr/local/bin/estaleiro",
			"apt-repositories",
			"--output=/keys.yml",
		}, append(keys, repositories...)...)),
		estaleiroSourceMount(),
	).Root()

	opts := []llb.RunOption{
		llb.Args(append([]string{
			"/usr/local/bin/estaleiro",
			"apt-packages",
			"--output=/pkgs.yml",
			"--debs=/var/lib/estaleiro/debs",
		}, packages...)),
	}

	opts = append(opts, estaleiroSourceMount(),
		llb.AddMount(
			"/var/lib/apt/lists",
			repositoriesState,
			llb.SourcePath("/var/lib/apt/lists"),
		),
		llb.AddMount(
			"/etc/apt/sources.list",
			repositoriesState,
			llb.SourcePath("/etc/apt/sources.list"),
		),
		llb.AddMount(
			"/etc/ssl/certs/ca-certificates.crt",
			repositoriesState,
			llb.SourcePath("/etc/ssl/certs/ca-certificates.crt"),
		),
	)

	if len(keys) > 0 {
		opts = append(opts, llb.AddMount(
			"/etc/apt/trusted.gpg",
			repositoriesState,
			llb.SourcePath("/etc/apt/trusted.gpg"),
		))
	}

	packagesState := fs.Run(opts...).Root()
	newBom = copy(repositoriesState, "/keys.yml", newBom, "/keys.yml")
	newBom = copy(packagesState, "/pkgs.yml", newBom, "/pkgs.yml")

	newFs = installPackages(newFs, packagesState, packages)

	return newFs, newBom
}

// installPackages installs debian packages that are provided from a given state
// `packages`.
//
func installPackages(base, pkgsFs llb.State, packages []string) llb.State {
	packagesMount := llb.AddMount(
		"/var/lib/estaleiro/debs",
		pkgsFs,
		llb.SourcePath("/var/lib/estaleiro/debs"),
	)

	run := base.Run(
		llb.Args(append([]string{
			"/usr/local/bin/estaleiro",
			"apt-install",
			"--debs=/var/lib/estaleiro/debs",
		}, packages...)),
		estaleiroSourceMount(),
		packagesMount,
	)

	return run.Root()
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

// given a tarball name and a file location within that tarball, finds out the `VCS`.
//
func getFileVCSInfo(cfg *config.Config, name, file string) *config.VCS {
	for _, t := range cfg.Tarballs {
		if t.Name != name {
			continue
		}

		for _, f := range t.SourceFiles {
			match, err := path.Match(f.Location, file)
			if err != nil {
				panic(errors.Wrapf(err, "invalid glob %s", f.Location))
			}

			if !match {
				continue
			}

			vcs := f.VCS
			return &vcs
		}
	}

	for _, t := range cfg.Steps {
		if t.Name != name {
			continue
		}

		for _, f := range t.SourceFiles {
			match, err := path.Match(f.Location, file)
			if err != nil {
				panic(errors.Wrapf(err, "invalid glob %s", f.Location))
			}

			if !match {
				continue
			}

			vcs := f.VCS
			return &vcs
		}
	}

	return nil
}

func tarballFiles(fs, bom llb.State, cfg *config.Config) (newFs llb.State, newBom llb.State, err error) {
	newFs, newBom = fs, bom

	// gather a list of the files that we're dealing with - those coming
	// from tarballs.
	//
	files := []config.File{}
	for _, file := range cfg.Image.Files {
		if file.FromTarball == nil {
			continue
		}

		files = append(files, file)
	}

	// maps `tarballName` to a list of files inside it that should be unpacked.
	//
	extractionsMap := map[string][]string{}
	for _, file := range files {
		filesToExtract, _ := extractionsMap[file.FromTarball.TarballName]
		extractionsMap[file.FromTarball.TarballName] = append(filesToExtract, file.FromTarball.Path)
	}

	// create the states where the files of each tarball will be extracted to
	//
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

	// gather VCS info in a per-file manner, and then save into the `bom` state in a file.
	//
	fileSourceMapping := make(map[string]bomfs.FileSource, len(files))
	for _, file := range files {
		// for that specific file in a given tarball, gather the source code info
		fileVCS := getFileVCSInfo(cfg, file.FromTarball.TarballName, file.FromTarball.Path)
		if fileVCS == nil {
			err = errors.Errorf("file with destination %s and source %s not declared in any tarball block",
				file.Destination, file.FromTarball.TarballName)
			return
		}

		fileSourceMapping[file.Destination] = bomfs.FileSource{
			Origin: bomfs.FileOrigin{
				Tarball: file.FromTarball.TarballName,
				Path:    file.FromTarball.Path,
			},
			VCS: *fileVCS,
		}
	}

	res, err := yaml.Marshal(bomfs.NewFileSourcesV1(fileSourceMapping))
	if err != nil {
		panic(err)
	}

	// save the sources info in the `bom` state
	//
	newBom = newBom.File(llb.Mkfile("/tarballs.yml", 0755, res))

	// copy the files to the fs state
	//
	for _, file := range files {
		tarballSourceState, found := tarballStateMap[file.FromTarball.TarballName]
		if !found {
			err = errors.Errorf("couldn't find tarball source state for tarball %s",
				file.FromTarball.TarballName)
			return
		}

		newFs = copy(tarballSourceState, file.FromTarball.Path, newFs, file.Destination)
	}

	return
}

func runAndCopyFromSteps(fs, bom llb.State, cfg *config.Config, dockerfileMapping map[string][]byte) (newFs, newBom llb.State, err error) {
	newFs, newBom = fs, bom

	// gather the config.Files that matter
	files := []config.File{}
	for _, file := range cfg.Image.Files {
		if file.FromStep == nil {
			continue
		}

		files = append(files, file)
	}

	fileSourceMapping := make(map[string]bomfs.FileSource, len(files))
	for _, file := range files {
		fileVCS := getFileVCSInfo(cfg, file.FromStep.StepName, file.FromStep.Path)
		if fileVCS == nil {
			err = errors.Errorf("file with destination %s and source %s not declared in any step block",
				file.Destination, file.FromStep.StepName)
		}

		fileSourceMapping[file.Destination] = bomfs.FileSource{
			Origin: bomfs.FileOrigin{
				Step: file.FromStep.StepName,
				Path: file.FromStep.Path,
			},
			VCS: *fileVCS,
		}

	}

	res, err := yaml.Marshal(bomfs.NewFileSourcesV1(fileSourceMapping))
	if err != nil {
		panic(err)
	}

	// save the sources info in the `bom` state
	//
	newBom = newBom.File(llb.Mkfile("/steps.yml", 0755, res))

	for _, file := range files {
		newFs, newBom, err = copyFileFromStep(newFs, newBom, cfg, file, dockerfileMapping)
		if err != nil {
			// TODO better error
			return
		}
	}

	return
}

func copyFileFromStep(
	fs, bom llb.State, cfg *config.Config, file config.File, dockerfileMapping map[string][]byte,
) (
	newFs, newBom llb.State, err error,
) {
	newFs, newBom = fs, bom

	// get the config's step definition that the file refers to.
	//
	configStep := getStepFromConfig(cfg, file.FromStep.StepName)
	if configStep == nil {
		err = errors.Errorf("referenced step %s not declared",
			file.FromStep.StepName)
		return
	}

	var step llb.State

	step, err = addImageBuildStep(configStep, dockerfileMapping[configStep.Dockerfile])
	if err != nil {
		err = errors.Wrapf(err,
			"failed to add step to image building process")
		return
	}

	newFs = copy(step, file.FromStep.Path, newFs, file.Destination)
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

func addImageBuildStep(step *config.Step, dockerfileContent []byte) (state llb.State, err error) {
	var stepState *llb.State

	buildContext := llb.Local("context")
	if step.Context != "" {
		buildContext = copy(
			llb.Local("context"),
			path.Join(step.Context, "*"),
			llb.Scratch(),
			"/")
	}

	caps := pb.Caps.CapSet(pb.Caps.All())
	stepState, _, err = dockerfile.Dockerfile2LLB(
		context.TODO(), dockerfileContent, dockerfile.ConvertOpt{
			Target:       step.Target,
			BuildContext: &buildContext,
			LLBCaps:      &caps,
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

// TODO - fix this to work on `docker build`s too (this relies on the fact that
// we can read from the fs), which is not really true - in a `docker build` (or
// any use of the gateway frontend), all you have is contexts (that you can
// leverage within the LLB only).
//
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
