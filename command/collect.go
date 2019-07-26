package command

import (
	"fmt"
	"io"
	"os"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
)

type collectCommand struct {
	DpkgStatusFilename string `long:"filename" default:"/var/lib/dpkg/status" descritpion:"path to dpkg status file ('-' for stdin)"`
}

func (c *collectCommand) Execute(args []string) (err error) {
	var (
		f        io.Reader
		packages []dpkg.Package
	)

	if c.DpkgStatusFilename == "-" {
		f = os.Stdin
	} else {
		var file *os.File

		file, err = os.Open(c.DpkgStatusFilename)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to open dpkg status file at %s", c.DpkgStatusFilename)
			return
		}
		defer file.Close()
		f = file
	}

	scanner := dpkg.NewScanner(f)
	packages, err = scanner.ScanAll()
	if err != nil {
		err = errors.Wrapf(err,
			"failed scanning packages from dpkg status file %s",
			c.DpkgStatusFilename)
		return
	}

	fmt.Println(packages)

	return
}
