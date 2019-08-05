package fs

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func Unarchive(ctx context.Context, tarball, dest string) (desc UnarchivedTarball, err error) {
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

func GatherUnarchivedFiles(
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
