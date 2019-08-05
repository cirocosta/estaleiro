package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type extractCommand struct {
	Destination string   `long:"destination" required:"true" description:"where to unarchive files to"`
	Files       []string `long:"file"        required:"true" description:"file from the tarball that we want to consume"`
	Output      string   `long:"output"      required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
	Tarball     string   `long:"tarball"     required:"true" description:"filepath to tarball to extract"`
}

type UnarchivedTarball struct {
	// Path to the original tarball
	//
	Path string `yaml:"path"`

	// Digest of the original tarball
	//
	Digest string `yaml:"digest"`

	// UnarchivedLocation corresponds to where the contents of the tarball
	// where unarchived to.
	//
	UnarchivedLocation string `yaml:"unarchived_location"`
}

type ExtractedFile struct {
	Name              string            `yaml:"name"`
	Path              string            `yaml:"path"`
	Digest            string            `yaml:"digest"`
	UnarchivedTarball UnarchivedTarball `yaml:"from_tarball"`
}

type FilesV1 struct {
	Kind string          `yaml:"kind"`
	Data []ExtractedFile `yaml:"data"`
}

func NewFilesV1(extractedFiles []ExtractedFile) FilesV1 {
	return FilesV1{
		Kind: "files/v1",
		Data: extractedFiles,
	}
}

func extract(ctx context.Context, tarball, dest string) (desc UnarchivedTarball, err error) {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		digest, err := dpkg.ComputeSHA256(tarball)
		if err != nil {
			return errors.Wrapf(err,
				"failed computing digest for tarball %s", tarball)
		}

		desc = UnarchivedTarball{
			Path:               tarball,
			Digest:             "sha256:" + digest,
			UnarchivedLocation: dest,
		}

		return nil
	})

	err = archiver.Unarchive(tarball, dest)
	if err != nil {
		err = errors.Wrapf(err,
			"failed unarchiving %s into %s", tarball, dest)
		return
	}

	err = eg.Wait()
	return

}

func gatherExtractedFiles(
	ctx context.Context, files []string, dest string, tarball UnarchivedTarball,
) (
	extracted []ExtractedFile, err error,
) {
	extracted = make([]ExtractedFile, len(files))
	eg, ctx := errgroup.WithContext(ctx)

	for idx, file := range files {
		idx, file := idx, file

		eg.Go(func() error {
			src := path.Join(tarball.UnarchivedLocation, file)
			dst := path.Join(dest, file)

			digest, err := dpkg.ComputeSHA256(src)
			if err != nil {
				return errors.Wrapf(err,
					"failed computing digest for file %s", src)
			}

			err = os.MkdirAll(filepath.Dir(file), 0755)
			if err != nil {
				return errors.Wrapf(err,
					"failed creating directory structure (%s) for unpacked files",
					filepath.Dir(file))

			}

			err = os.Rename(src, dst)
			if err != nil {
				return errors.Wrapf(err,
					"failed moving file from %s to %s", src, dst)
			}

			extracted[idx] = ExtractedFile{
				Name:              file,
				Digest:            "sha256:" + digest,
				Path:              dst,
				UnarchivedTarball: tarball,
			}

			return nil
		})

	}

	err = eg.Wait()
	if err != nil {
		return
	}

	return

}

// TODO - compute tarball's digest too (can be done concurrently)
// TODO - make digest computation concurrent
//
func (c *extractCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	dir, err := ioutil.TempDir(c.Destination, "estailero-tar")
	if err != nil {
		err = errors.Wrapf(err,
			"failed creating directory for tar extraction")
		return
	}

	defer os.RemoveAll(dir)

	tarballDesc, err := extract(ctx, c.Tarball, dir)
	if err != nil {
		err = errors.Wrapf(err,
			"failed extracting tarball %s to %s", c.Tarball, dir)
		return
	}

	extracted, err := gatherExtractedFiles(ctx, c.Files, c.Destination, tarballDesc)
	if err != nil {
		err = errors.Wrapf(err,
			"failed gathering files from extracted tarball %s", c.Tarball)
		return
	}

	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	b, err := yaml.Marshal(NewFilesV1(extracted))
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
