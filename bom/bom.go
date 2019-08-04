package bom

import (
	"time"

	"encoding/json"
	"github.com/cirocosta/estaleiro/dpkg"
	"gopkg.in/yaml.v3"
)

// Bom represents the final bill of materials.
//
type Bom struct {

	// ProductName is the name of the product that the bill of materials is
	// describing.
	//
	ProductName string `yaml:"product_name"`

	// Version corresponds to the version of the format that the BOM adheres
	// to, not the version of the product.
	//
	//
	Version string `yaml:"version"`

	GeneratedAt time.Time `yaml:"generated_at"`

	BaseImage BaseImage `yaml:"base_image"`

	ChangeSet struct {
		Files    []File         `yaml:"files"`
		Packages []dpkg.Package `yaml:"packages"`
	} `yaml:"changeset"`
}

type Source struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
	Uri     string `yaml:"uri"`
}

type File struct {
	// Path to the file within the final container image.
	//
	Path string `yaml:"path"`

	// Digest corresponds to the computation of a given digest algorithm on
	// the bits in that path.
	//
	// It takes the form of `alg:<result>`.
	//
	// For instance: `sha256:blablabla`
	//
	Digest string `yaml:"digest"`

	// TODO
	// FromTarball FromTarball `yaml:"from_tarball"`

	// Source brings information related to the actual code that produced
	// the thing pointed at by the `path`.
	//
	// For instance, if a binary `bin` was compiled from a given git
	// repository, `source` would point to that git repository.
	//
	Source Source `yaml:"source"`
}

type BaseImage struct {
	Name   string `yaml:"name"`
	Digest string `yaml:"digest"`

	OS       string `yaml:"os"`
	Version  string `yaml:"version"`
	Codename string `yaml:"codename"`

	// Packages lists all of the packages that were present there in the
	// base image, before any user packages were added.
	//
	Packages []dpkg.Package `yaml:"packages"`
}

func (b Bom) ToJSON() (res []byte) {
	var err error

	res, err = json.Marshal(&b)
	if err != nil {
		panic(err)
	}

	return
}

func (b Bom) ToYAML() (res []byte) {
	var err error

	res, err = yaml.Marshal(&b)
	if err != nil {
		panic(err)
	}

	return
}
