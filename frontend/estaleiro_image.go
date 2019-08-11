// +build !local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

const (
	imageName = "cirocosta/estaleiro@sha256:703046892748abd2aa36310fc2f7f92c66ee278aeef94c1b09c844178f7b7b7e"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Image(imageName),
		llb.SourcePath("/usr/local/bin/estaleiro"),
	)
}
