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
	BomDestination string            `long:"bom" short:"b" description:"where to save the bill of materials to"`
	Filename       string            `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	Variables      map[string]string `long:"var" short:"v" description:"variables to interpolate"`
}

func (c *llbCommand) Execute(args []string) (err error) {
	color.NoColor = false

	cfg, err := config.ParseFile(c.Filename, c.Variables)
	if err != nil {

		diagsErr, ok := errors.Cause(err).(hcl.Diagnostics)
		if ok {
			fmt.Fprintln(os.Stderr, config.PrettyDiagnosticFile(c.Filename, diagsErr[0]))
		}

		err = errors.Wrapf(err, "failed to parse config file %s", c.Filename)
		return
	}

	state, _, bom, err := frontend.ToLLB(context.TODO(), cfg)
	if err != nil {
		err = errors.Wrapf(err, "failed to convert to llb")
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

	definition, err := state.Marshal(llb.LinuxAmd64)
	if err != nil {
		err = errors.Wrap(err, "marshaling llb state")
		return
	}

	llb.WriteTo(definition, os.Stdout)

	return
}
