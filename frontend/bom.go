package frontend

import (
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
	Files []File `yaml:"files"`
}

func (b Bom) ToYAML() (res []byte) {
	var err error

	res, err = yaml.Marshal(&b)
	if err != nil {
		panic(err)
	}

	return
}
