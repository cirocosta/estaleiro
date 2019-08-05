package fs

import (
	"github.com/cirocosta/estaleiro/config"
)

type FileSourcesV1 struct {
	Kind string `yaml:"kind"`

	// maps file location (in the final image) to VCS info.
	//
	Data map[string]config.VCS `yaml:"data"`
}

func NewFileSourcesV1(vcsMapping map[string]config.VCS) FileSourcesV1 {
	return FileSourcesV1{
		Kind: "filesources/v1",
		Data: vcsMapping,
	}
}
