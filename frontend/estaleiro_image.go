// +build !local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

const (
	imageName = "cirocosta/estaleiro@sha256:de56c2fecf9a0197d3ac25bd6df6fe7287e8a137ead48f100fc93b80bf8c165e"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Image(imageName),
		llb.SourcePath("/usr/local/bin/estaleiro"),
	)
}
