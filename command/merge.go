package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cirocosta/estaleiro/bom"
	bomfs "github.com/cirocosta/estaleiro/bom/fs"
	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type mergeCommand struct {
	Directory string `long:"directory" required:"true" description:"where bill of material files live"`
	Output    string `long:"output"    required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

func (c *mergeCommand) Execute(args []string) (err error) {
	glob := filepath.Join(c.Directory, "*.yml")

	files, err := filepath.Glob(glob)
	if err != nil {
		err = errors.Wrapf(err,
			"failed listing yml files with glob %s", glob)
		return
	}

	var (
		filesv1         = []bomfs.FilesV1{}
		filesourcesv1   = []bomfs.FileSourcesV1{}
		osPackagesv1    = bomfs.PackagesV1{}
		addedPackagesv1 = bomfs.PackagesV1{}
		osReleasev1     = bomfs.OsReleaseV1{}
		metav1          = bomfs.MetaV1{}
	)

	for _, file := range files {
		obj, err := readTypedBomFile(file)
		if err != nil {
			return err
		}

		switch v := obj.(type) {
		case *bomfs.FilesV1:
			filesv1 = append(filesv1, *v)
		case *bomfs.FileSourcesV1:
			filesourcesv1 = append(filesourcesv1, *v)
		case *bomfs.OsReleaseV1:
			osReleasev1 = *v
		case *bomfs.MetaV1:
			metav1 = *v
		case *bomfs.PackagesV1:
			if v.Data.Initial {
				osPackagesv1 = *v
			} else {
				addedPackagesv1 = *v
			}
		}
	}

	materials := bom.Bom{}

	materials.ProductName = metav1.Data.ProductName
	materials.BaseImage.CanonicalName = metav1.Data.Image
	materials.BaseImage.OS = osReleasev1.Data.OS
	materials.BaseImage.Version = osReleasev1.Data.Version
	materials.BaseImage.Codename = osReleasev1.Data.Codename
	materials.BaseImage.Packages = toBomPackages(osPackagesv1.Data.Packages)
	materials.ChangeSet.Packages = toBomPackages(addedPackagesv1.Data.Packages)
	materials.ChangeSet.Files = toBomFiles(filesv1, filesourcesv1)

	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	fmt.Fprintf(writer, "%s", materials.ToYAML())

	return
}

func toBomFiles(files []bomfs.FilesV1, sources []bomfs.FileSourcesV1) (res []bom.File) {
	extracted := []bomfs.ExtractedFile{}
	for _, f := range files {
		extracted = append(extracted, f.Data...)
	}

	imageFiles := map[string]bomfs.FileSource{}
	for _, s := range sources {
		for f, fileSource := range s.Data {
			imageFiles[f] = fileSource
		}
	}

	res = make([]bom.File, len(imageFiles))
	idx := 0

	for imageFileName, fileSource := range imageFiles {
		extractedFile := getExtractedFile(extracted, fileSource.Origin)

		res[idx] = bom.File{
			Path:   imageFileName,
			Digest: extractedFile.Digest,
			Source: bom.Source{
				Git: &bom.GitSource{
					RepositoryUri: fileSource.VCS.Repository,
					Ref:           fileSource.VCS.Ref,
				},
			},
		}

		idx++
	}

	return
}

func getExtractedFile(files []bomfs.ExtractedFile, origin bomfs.FileOrigin) (res bomfs.ExtractedFile) {
	for _, file := range files {
		// this seems fragile
		if file.UnarchivedTarball.Path != filepath.Join("/src", origin.Tarball) {
			continue
		}

		if file.Name != origin.Path {
			continue
		}

		res = file
		return
	}

	return
}

func toBomPackages(orig []dpkg.Package) (res []bom.Package) {
	res = make([]bom.Package, len(orig))
	emptyLocation := dpkg.AptDebLocation{}

	for idx, pkg := range orig {
		sources := make([]bom.ExternalResource, len(pkg.Source))

		for idx, source := range pkg.Source {
			sources[idx] = bom.ExternalResource{
				Uri:    source.URI,
				Digest: "md5:" + source.MD5sum,
				Name:   source.Name,
			}
		}

		location := bom.ExternalResource{}
		if pkg.Location != emptyLocation {
			location = bom.ExternalResource{
				Uri:    pkg.Location.URI,
				Digest: "md5:" + pkg.Location.MD5sum,
				Name:   pkg.Location.Name,
			}
		}

		res[idx] = bom.Package{
			Name:          pkg.Name,
			Version:       pkg.Version,
			SourcePackage: pkg.SourcePackage,
			Architecture:  pkg.Architecture,
			Location:      location,
			Sources:       sources,
		}
	}

	return
}

func readTypedBomFile(file string) (res interface{}, err error) {
	var (
		f *os.File
		b []byte
	)

	f, err = os.Open(file)
	if err != nil {
		err = errors.Wrapf(err,
			"failed opening file %s", file)
		return
	}

	defer f.Close()

	b, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.Wrapf(err,
			"failed reading contents from file %s", file)
		return
	}

	wrapper := struct {
		Kind string `yaml:"kind"`
	}{}

	err = yaml.Unmarshal(b, &wrapper)
	if err != nil {
		err = errors.Wrapf(err,
			"failed unmarshalling data into wrapper for type discoverabiility", err)
		return
	}

	switch wrapper.Kind {
	case bomfs.FilesV1Kind:
		res = &bomfs.FilesV1{}
	case bomfs.FileSourcesV1Kind:
		res = &bomfs.FileSourcesV1{}
	case bomfs.OsReleaseV1Kind:
		res = &bomfs.OsReleaseV1{}
	case bomfs.PackagesV1Kind:
		res = &bomfs.PackagesV1{}
	case bomfs.MetaV1Kind:
		res = &bomfs.MetaV1{}
	default:
		err = errors.Wrapf(err,
			"unexpected kind %s", wrapper.Kind)
		return
	}

	err = yaml.Unmarshal(b, res)
	if err != nil {
		err = errors.Wrapf(err,
			"failed unmarshalling data into "+wrapper.Kind, err)
		return
	}

	return
}
