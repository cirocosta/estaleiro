// +build !local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

const (
	imageName = "cirocosta/estaleiro@sha256:40dbe8b0cc58dca2d71a2f1c60f6aa8573facf3ae21e4fc7fd3eca2bbeecada9"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Image(imageName),
		llb.SourcePath("/usr/local/bin/estaleiro"),
	)
}
