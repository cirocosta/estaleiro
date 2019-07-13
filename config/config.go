package config

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/pkg/errors"
)

type Config struct {
	Image Image `hcl:"image,block"`
}

type Image struct {
	BaseImage BaseImage `hcl:"base_image"`
}

type BaseImage struct {
	Name      string `hcl:"name"`
	Reference string `hcl:"ref"`
}

func ParseFile(filename string) (cfg *Config, err error) {
	var (
		file    *os.File
		content []byte
	)

	file, err = os.Open(filename)
	if err != nil {
		return
	}

	defer file.Close()

	content, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return Parse(content, filename)
}

func Parse(content []byte, filename string) (cfg *Config, err error) {
	f, diag := hclsyntax.ParseConfig(content, filename, hcl.Pos{})
	if diag.HasErrors() {
		err = errors.Wrapf(diag, "failed to parse")
		return
	}

	cfg = new(Config)

	diag = gohcl.DecodeBody(f.Body, nil, cfg)
	if diag.HasErrors() {
		err = errors.Wrapf(diag, "failed to decode")
		return
	}

	return
}
