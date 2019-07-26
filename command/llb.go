package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cirocosta/estaleiro/bom"
	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

type llbCommand struct {
	BomDestination string            `long:"bom" short:"b" description:"where to save the bill of materials to"`
	Filename       string            `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	Variables      map[string]string `long:"var" short:"v" description:"variables to interpolate"`
}

func HCLToLLB(filename string, variables map[string]string) (definition *llb.Definition, bom bom.Bom, err error) {
	cfg, err := config.ParseFile(filename, variables)
	if err != nil {
		diagsErr, ok := errors.Cause(err).(hcl.Diagnostics)
		if ok {
			fmt.Fprintln(os.Stderr, config.PrettyDiagnosticFile(filename, diagsErr[0]))
		}

		err = errors.Wrapf(err, "failed to parse config file %s", filename)
		return
	}

	var state llb.State
	state, _, bom, err = frontend.ToLLB(context.TODO(), cfg)
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

func init() {
	color.NoColor = false
}

func (c *llbCommand) Execute(args []string) (err error) {
	def, bom, err := HCLToLLB(c.Filename, c.Variables)
	if err != nil {
		return
	}

	if c.BomDestination != "" {
		err = ioutil.WriteFile(c.BomDestination, bom.ToYAML(), 0644)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to write bill of materials to file %s",
				c.BomDestination)
			return
		}
	}

	llb.WriteTo(def, os.Stdout)

	return
}
