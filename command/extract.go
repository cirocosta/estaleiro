package command

import (
	"os"
	"path"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type extractCommand struct {
	Tarball     string   `long:"tarball"     required:"true" description:"filepath to tarball to extract"`
	Files       []string `long:"file"        required:"true" description:"file from the tarball that we want to consume"`
	Destination string   `long:"destination" required:"true" description:"where to unarchive files to"`
}

func mapFromSlice(items []string) (res map[string]interface{}) {
	res = make(map[string]interface{}, len(items))

	for _, item := range items {
		res[item] = nil
	}

	return
}

func (c *extractCommand) Execute(args []string) (err error) {
	files := mapFromSlice(c.Files)

	err = archiver.Unarchive(c.Tarball, c.Destination)
	if err != nil {
		err = errors.Wrapf(err,
			"failed unarchiving %s into %s", c.Tarball, c.Destination)
		return
	}

	// TODO this could be performed concurrently
	for _, file := range c.Files {
		filepath := path.Join(c.Destination, file)

		digest, err := dpkg.ComputeSHA256(filepath)
		if err != nil {
			return
		}

	}

	// unarchive everything somewhere
	// walk through the unarchived dir
	// compute the SHA's of the files that we're interested in

	archiver.Walk(c.Tarball, func(f archiver.File) (err error) {
		_, found := files[f.Name()]
		if !found {
			return
		}
	})

	// walk inside a given tarball
	// extract the files that match our list of files
	//

	return
}
