package frontend

import (
	"context"
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
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	copyImage = "docker.io/library/alpine:latest@sha256:1072e499f3f655a032e88542330cf75b02e7bdf673278f701d7ba61629ee3ebe"
)

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

func ToLLB(cfg *config.Config) (state llb.State, bom Bom, err error) {
	bom.Version = "v0.0.1"
	bom.GeneratedAt = time.Now()

	// TODO - consider tag provided
	canonicalName, err := resolveImage(context.TODO(), cfg.Image.BaseImage.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to resolve digest for %s when preparing llb", cfg.Image.BaseImage.Name)
		return
	}

	bom.BaseImage = BaseImage{
		Name:   canonicalName.Name(),
		Digest: canonicalName.Digest().String(),
	}

	state = llb.Image(canonicalName.String())

	for _, file := range cfg.Image.Files {
		if file.FromStep == nil {
			continue
		}

		// retrieve file from step
		//

		configStep := getStepFromConfig(cfg, file.FromStep.StepName)
		if configStep == nil {
			err = errors.Errorf("referenced step %s not declared",
				file.FromStep.StepName)
			return
		}

		// the file has a `path`
		// -- make sure that the `path` matches a `source_file`
		//    in the `step` definition

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
		step, err = addStep(configStep)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to add step to image building process")
			return
		}

		state = copy(step, file.FromStep.Path, state, file.Destination)
	}

	// write bill of materials somewhere?

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

// copy copies files between 2 states using cp until there is no copyOp
//
func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	cpImage := llb.Image(copyImage)

	cp := cpImage.Run(llb.Shlexf("cp -a /src%s /dest%s", srcPath, destPath))
	cp.AddMount("/src", src, llb.Readonly)

	return cp.AddMount("/dest", dest)
}

func addStep(step *config.Step) (state llb.State, err error) {
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
