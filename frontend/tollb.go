package frontend

import (
	"context"

	"github.com/cirocosta/estaleiro/config"
	"github.com/moby/buildkit/client/llb"
	dockerfile "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/pkg/errors"
)

const (
	copyImage = "docker.io/library/alpine:latest@sha256:1072e499f3f655a032e88542330cf75b02e7bdf673278f701d7ba61629ee3ebe"
)

func ToLLB(cfg *config.Config) (state llb.State, err error) {
	state = llb.Image(cfg.Image.BaseImage.Name)

	for _, file := range cfg.Image.Files {
		if file.FromStep != nil {
			configStep := getStepFromConfig(cfg, file.FromStep.StepName)
			if configStep == nil {
				err = errors.Errorf("referenced step %s not declared", file.FromStep.StepName)
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
	}

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

func addStep(*config.Step) (state llb.State, err error) {
	var stepState *llb.State

	stepState, _, err = dockerfile.Dockerfile2LLB(context.TODO(), []byte(`
FROM busybox
RUN echo "lol"`), dockerfile.ConvertOpt{})
	if err != nil {
		err = errors.Wrapf(err,
			"failed to convert dockerfile to llb")
		return
	}

	state = *stepState
	return
}
