package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

type llbCommand struct {
	Filename  string            `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	Variables map[string]string `long:"var" short:"v" description:"variables to interpolate"`
}

func HCLToLLB(filename string, variables map[string]string) (definition *llb.Definition, err error) {
	cfg, err := config.ParseFile(filename, variables)
	if err != nil {
		diagsErr, ok := errors.Cause(err).(hcl.Diagnostics)
		if ok {
			fmt.Fprintln(os.Stderr, config.PrettyDiagnosticFile(filename, diagsErr[0]))
		}

		err = errors.Wrapf(err, "failed to parse config file %s", filename)
		return
	}

	mapping := make(map[string][]byte)
	for _, step := range cfg.Steps {
		_, found := mapping[step.Dockerfile]
		if found {
			continue
		}

		mapping[step.Dockerfile], err = readFile(step.Dockerfile)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to read dockerfile from local directory")
			return
		}
	}

	var state llb.State
	state, _, err = frontend.ToLLB(context.TODO(), cfg, mapping)
	if err != nil {
		err = errors.Wrapf(err, "failed to convert to llb")
		return
	}

	definition, err = state.Marshal(llb.LinuxAmd64)
	if err != nil {
		err = errors.Wrap(err, "marshaling llb state")
		return
	}

	return
}

func readFile(filename string) (content []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open file %s", filename)
		return
	}

	defer f.Close()

	content, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.Wrapf(err,
			"failed reading contents of file %s", filename)
		return
	}

	return
}

func init() {
	color.NoColor = false
}

func (c *llbCommand) Execute(args []string) (err error) {
	def, err := HCLToLLB(c.Filename, c.Variables)
	if err != nil {
		return
	}

	llb.WriteTo(def, os.Stdout)

	return
}
