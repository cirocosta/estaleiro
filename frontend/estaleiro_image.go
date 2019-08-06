// +build !local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

const (
	imageName = "cirocosta/estaleiro@sha256:a8e0fb0557e133b1cf74e1b7a6bb160e2070f92a659a0e1dc7e2036a8e53d4bd"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Image(imageName),
		llb.SourcePath("/usr/local/bin/estaleiro"),
	)
}
