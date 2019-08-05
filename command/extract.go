package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type extractCommand struct {
	Destination string   `long:"destination" required:"true" description:"where to unarchive files to"`
	Files       []string `long:"file"        required:"true" description:"file from the tarball that we want to consume"`
	Output      string   `long:"output"      required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
	Tarball     string   `long:"tarball"     required:"true" description:"filepath to tarball to extract"`
}

func (c *extractCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	dir, err := ioutil.TempDir(c.Destination, "estailero-tar")
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating directory for tar extraction")
		return
	}

	defer os.RemoveAll(dir)

	tarballDesc, err := bomfs.Unarchive(ctx, c.Tarball, dir)
	if err != nil {
		err = errors.Wrapf(err,
			"failed extracting tarball %s to %s", c.Tarball, dir)
		return
	}

	extracted, err := bomfs.GatherUnarchivedFiles(ctx, c.Files, c.Destination, tarballDesc)
	if err != nil {
		err = errors.Wrapf(err,
			"failed gathering files from extracted tarball %s", c.Tarball)
		return
	}

	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	b, err := yaml.Marshal(bomfs.NewFilesV1(extracted))
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(writer, "%s", string(b))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}
