package command

import (
	"context"
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type aptCommand struct {
	Output string `long:"output" required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

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

func (c *aptCommand) Execute(args []string) (err error) {
	pkgs, err := dpkg.InstallPackages(context.TODO(), args)
	if err != nil {
		err = errors.Wrapf(err,
			"failed installing packages")
		return
	}

	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	res, err := yaml.Marshal(NewPackagesV1(false, pkgs))
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(writer, "%s", string(res))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}
