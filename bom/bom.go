// Package bom defines the final bill of materials (BOM).
//
package bom

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

const Version = "v0.0.1"

// Bom represents the final consolidated bill of materials.
//
type Bom struct {

	// ProductName is the name of the product that the bill of materials is
	// describing.
	//
	ProductName string `yaml:"product_name"`

	// Protocol corresponds to the version of the format that the BOM adheres
	// to, not the version of the product.
	//
	//
	Protocol string `yaml:"protocol"`

	// BaseImage aggregates information about the base image that provided
	// the initial rootfs for the final container image.
	//
	BaseImage BaseImage `yaml:"base_image"`

	// ChangeSet represents the modifications performed on top of the base image.
	//
	ChangeSet struct {
		Files    []File    `yaml:"files"`
		Packages []Package `yaml:"packages"`
	} `yaml:"changeset"`
}

type BaseImage struct {

	// CanonicalName is the full name of the base image, containing the
	// repository, name and digest.
	//
	CanonicalName string `yaml:"canonical_name"`

	// OS-release information

	OS       string `yaml:"os"`
	Version  string `yaml:"version"`
	Codename string `yaml:"codename"`

	// Packages lists all of the packages that were present there in the
	// base image, before any user packages were added.
	//
	// This helps disambiguate packages that were added on top of the base
	// image and those that were there by default.
	//
	Packages []Package `yaml:"packages"`
}

// Package represents a debian package, optionally containing information about
// how to have it retrieved (including its sources).
//
type Package struct {
	Name          string             `yaml:"name"`
	Version       string             `yaml:"version"`
	SourcePackage string             `yaml:"source_package"`
	Architecture  string             `yaml:"architecture"`
	Location      ExternalResource   `yaml:"location,omitempty"`
	Sources       []ExternalResource `yaml:"sources,omitempty"`
}

// ExternalResource indicates an external resource that can be retrieved from a
// given URI and verified by checking its digest.
//
type ExternalResource struct {
	Uri    string `yaml:"uri"`
	Digest string `yaml:"digest"`

	// Name is the name of the resource that would be downloaded.
	//
	// For instance, `my-tarball.tgz`.
	//
	Name string `yaml:"name,omitempty"`
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

	// Source brings information related to the actual code that produced
	// the thing pointed at by the `path`.
	//
	// For instance, if a binary `bin` was compiled from a given git
	// repository, `source` would point to that git repository.
	//
	Source Source `yaml:"source"`
}

type Source struct {
	Git *GitSource `yaml:"git"`
}

type GitSource struct {

	// RepositoryUri is the URI that allows one to clone a git repository
	// that contains the source code that originates such file.
	//
	// Example: `https://github.com/cirocosta/estaleiro.git`
	//
	RepositoryUri string `yaml:"repository_uri"`

	// Ref is the git reference of the git repository where the source code
	// that originated the file exists.
	//
	// Examples:
	// - `v1.2.3` (tag)
	// - `57bf9d7e6fceee57b73ce97c9eeea5d30fd3125b` (commit SHA)
	// - `dev` (branch)
	//
	Ref string `yaml:"ref"`
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
