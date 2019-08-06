// +build local

package frontend

import (
	"github.com/moby/buildkit/client/llb"
)

func estaleiroSourceMount() llb.RunOption {
	return llb.AddMount(
		"/usr/local/bin/estaleiro",
		llb.Local("bin"),
		llb.SourcePath("estaleiro"),
	)
}
