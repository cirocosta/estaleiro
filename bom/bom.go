package bom

import (
	"time"

	"github.com/cirocosta/estaleiro/dpkg"
	"gopkg.in/yaml.v3"
)

// Bom represents the final bill of materials.
//
type Bom struct {
	Version     string    `yaml:"version"`
	GeneratedAt time.Time `yaml:"generated_at"`

	BaseImage BaseImage `yaml:"base_image"`

	Files    []File         `yaml:"added_files"`
	Packages []dpkg.Package `yaml:"added_packages"`
}

type Source struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
	Uri     string `yaml:"uri"`
}

type File struct {
	Path   string `yaml:"path"`
	Source Source `yaml:"source"`
}

type BaseImage struct {
	Name     string         `yaml:"name"`
	Digest   string         `yaml:"digest"`
	Packages []dpkg.Package `yaml:"packages"`
}

func (b Bom) ToYAML() (res []byte) {
	var err error

	res, err = yaml.Marshal(&b)
	if err != nil {
		panic(err)
	}

	return
}
