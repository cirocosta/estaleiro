package command

import (
	"context"
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"
	"gopkg.in/yaml.v3"
)

type aptCommand struct{}

func (c *aptCommand) Execute(args []string) (err error) {
	pkgs, err := dpkg.InstallPackages(context.TODO(), args)
	if err != nil {
		return
	}

	res, err := yaml.Marshal(pkgs)
	if err != nil {
		return
	}

	fmt.Println(string(res))

	return
}
