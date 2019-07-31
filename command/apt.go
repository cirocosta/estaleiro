package command

import (
	"context"
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type aptCommand struct{}

func (c *aptCommand) Execute(args []string) (err error) {
	pkgs, err := dpkg.InstallPackages(context.TODO(), args)
	if err != nil {
		err = errors.Wrapf(err,
			"failed installing packages")
		return
	}

	res, err := yaml.Marshal(pkgs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res))

	return
}
