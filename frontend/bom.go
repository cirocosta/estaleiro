package frontend

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
	Uri     string `yaml:"uri"`
}

type File struct {
	Path   string `yaml:"path"`
	Source Source `yaml:"source"`
}
type Bom struct {
	Version     string    `yaml:"version"`
	GeneratedAt time.Time `yaml:"generated_at"`

	BaseImage BaseImage `yaml:"base_image"`
	Files     []File    `yaml:"files"`
	Packages  []Package `yaml:"packages"`
}

type BaseImage struct {
	Name   string `yaml:"name"`
	Digest string `yaml:"digest"`
}

type Package struct {
	Name string `yaml:"name"`
}

func (b Bom) ToYAML() (res []byte) {
	var err error

	res, err = yaml.Marshal(&b)
	if err != nil {
		panic(err)
	}

	return
}
