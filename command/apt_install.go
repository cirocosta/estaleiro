package command

import (
	"context"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
)

type aptInstallCommand struct {
	Packages []string `short:"p" required:"true"`

	DebianPackagesDirectory string `long:"debs"   default:"/var/lib/estaleiro/debs"`
}

func (c *aptInstallCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	err = dpkg.ForceLocalSourcesList(ctx, c.DebianPackagesDirectory)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to force apt to use local repositories")
		return
	}

	err = dpkg.InstallDebianPackages(ctx, c.Packages)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install downloaded debian packages")
		return
	}

	err = dpkg.RemoveAptLists()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to remove apt repository listing after installation")
		return
	}

	return
}
