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

	res, err := yaml.Marshal(pkgs)
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(writer, "%+v", string(res))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}
