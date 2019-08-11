package command

import (
	"context"
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	bomfs "github.com/cirocosta/estaleiro/bom/fs"
)

type aptCommand struct {
	Keys         []string `long:"key" description:"additional keys to add to the keyring"`
	Repositories []string `long:"repository" description:"aditional apt repositories to add"`
	Output       string   `long:"output" required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

func (c *aptCommand) Execute(args []string) (err error) {
	pkgs, err := dpkg.InstallPackages(context.TODO(), c.Repositories, args)
	if err != nil {
		err = errors.Wrapf(err,
			"failed installing packages")
		return
	}

	w, err := writer(c.Output)
	if err != nil {
		return
	}

	res, err := yaml.Marshal(bomfs.NewPackagesV1(false, pkgs))
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, "%s", string(res))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}
