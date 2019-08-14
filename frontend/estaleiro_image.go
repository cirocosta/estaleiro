// +build !local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

const (
	imageName = "cirocosta/estaleiro@sha256:f6a1fae66bbc07929837c84749e22783df7394a7cfc3fadbb2c9e67b5057c5fd"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Image(imageName),
		llb.SourcePath("/usr/local/bin/estaleiro"),
	)
}
