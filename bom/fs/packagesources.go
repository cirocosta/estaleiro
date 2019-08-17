package fs

import (
	"github.com/cirocosta/estaleiro/config"
)

const (
	PackageSourcesV1Kind = "packagesources/v1"
)

type PackageOrigin struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type PackageSource struct {
	VCS    config.VCS    `yaml:",inline"`
	Origin PackageOrigin `yaml:"origin"`
}

type PackageSourcesV1 struct {
	Kind string `yaml:"kind"`

	// maps `package=verson` to VCS info.
	//
	Data map[string]PackageSource `yaml:"data"`
}

func NewPackageSourcesV1(mapping map[string]PackageSource) PackageSourcesV1 {
	return PackageSourcesV1{
		Kind: PackageSourcesV1Kind,
		Data: mapping,
	}
}
