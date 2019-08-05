package fs

import (
	"github.com/cirocosta/estaleiro/config"
)

const (
	FileSourcesV1Kind = "filesources/v1"
)

type FileOrigin struct {
	Tarball string `yaml:"tarball,omitempty"`
	Step    string `yaml:"step,omitempty"`
	Path    string `yaml:"path"`
}

type FileSource struct {
	VCS    config.VCS `yaml:",inline"`
	Origin FileOrigin `yaml:"origin"`
}

type FileSourcesV1 struct {
	Kind string `yaml:"kind"`

	// maps file location (in the final image) to VCS info.
	//
	Data map[string]FileSource `yaml:"data"`
}

func NewFileSourcesV1(vcsMapping map[string]FileSource) FileSourcesV1 {
	return FileSourcesV1{
		Kind: FileSourcesV1Kind,
		Data: vcsMapping,
	}
}
