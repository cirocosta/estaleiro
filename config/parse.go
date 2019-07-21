package config

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

func ParseFile(filename string, vars map[string]string) (cfg *Config, err error) {
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

	return Parse(content, filename, vars)
}

func Parse(content []byte, filename string, vars map[string]string) (cfg *Config, err error) {
	f, diag := hclsyntax.ParseConfig(content, filename, hcl.Pos{})
	if diag.HasErrors() {
		err = errors.Wrapf(diag, "failed to parse")
		return
	}

	cfg = new(Config)

	diag = gohcl.DecodeBody(f.Body, createEvalContext(vars), cfg)
	if diag.HasErrors() {
		err = errors.Wrapf(diag, "failed to decode")
		return
	}

	return
}

func createEvalContext(vars map[string]string) *hcl.EvalContext {
	var variables = map[string]cty.Value{}

	for key, value := range vars {
		variables[key] = cty.StringVal(value)
	}

	return &hcl.EvalContext{Variables: variables}
}
