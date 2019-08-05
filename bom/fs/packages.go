package fs

import (
	"github.com/cirocosta/estaleiro/dpkg"
)

type PackagesV1 struct {
	Kind string         `yaml:"kind"`
	Data PackagesV1Data `yaml:"data"`
}

type PackagesV1Data struct {
	Initial  bool           `yaml:"initial"`
	Packages []dpkg.Package `yaml:"packages"`
}

func NewPackagesV1(initial bool, pkgs []dpkg.Package) PackagesV1 {
	return PackagesV1{
		Kind: "packages/v1",
		Data: PackagesV1Data{
			Initial:  initial,
			Packages: pkgs,
		},
	}
}
