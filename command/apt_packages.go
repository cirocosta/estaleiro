package command

import (
	"context"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type aptPackagesCommand struct {
	Output                  string   `long:"output" default:"-"`
	DebianPackagesDirectory string   `long:"debs"            default:"/var/lib/estaleiro/debs"`
	Packages                []string `short:"p"              required:"true"`
}

func (c *aptPackagesCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	pkgs, err := retrieveDebianPackages(ctx, c.Packages, c.DebianPackagesDirectory)
	if err != nil {
		return
	}

	err = writePackagesBom(c.Output, pkgs)
	if err != nil {
		return
	}

	return
}

func retrieveDebianPackages(ctx context.Context, packages []string, dir string) (pkgs []dpkg.Package, err error) {
	logger.Info("retrieve-packages", lager.Data{"packages": packages, "dir": dir})

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating debian packages directory %s", dir)
		return
	}

	logger.Info("gather-deb-locations")
	locations, err := dpkg.GatherDebLocations(ctx, packages)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to install packages %v", packages)
		return
	}

	logger.Info("download-debian-packages")
	err = dpkg.DownloadDebianPackages(ctx, dir, locations)
	if err != nil {
		err = errors.Wrapf(err,
			"failed downloading packages %v", packages)
		return
	}

	logger.Info("create-packages")
	pkgs, err = dpkg.CreatePackages(ctx, dir, locations)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create packages bom")
		return
	}

	logger.Info("create-debian-repo-index", lager.Data{"dir": dir})
	err = dpkg.CreateDebianRepositoryIndex(dir, pkgs)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating debian repository index")
		return
	}

	return
}

func writePackagesBom(fname string, pkgs []dpkg.Package) (err error) {
	w, err := writer(fname)
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating writer for file %s", fname)
		return
	}

	b, err := yaml.Marshal(bomfs.NewPackagesV1(false, pkgs))
	if err != nil {
		err = errors.Wrapf(err,
			"failed marshalling bomfs packages")
		return
	}

	_, err = fmt.Fprintf(w, "%s", string(b))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", fname)
		return
	}

	return
}
