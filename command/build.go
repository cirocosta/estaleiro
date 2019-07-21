package command

import (
	"io/ioutil"
	"os"

	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

type buildCommand struct {
	BomDestination string `long:"bom" short:"b" description:"where to save the bill of materials to"`
	Filename       string `long:"filename" short:"f" required:"true" description:"file containing image definition"`
}

func (c *buildCommand) Execute(args []string) (err error) {
	cfg, err := config.ParseFile(c.Filename)
	if err != nil {
		err = errors.Wrapf(err, "failed ot parse config file %s", c.Filename)
		return
	}

	state, bom, err := frontend.ToLLB(cfg)
	if err != nil {
		err = errors.Wrapf(err, "failed to convert to llb")
		return
	}

	if c.BomDestination != "" {
		err = ioutil.WriteFile(c.BomDestination, bom.ToYAML(), 0755)
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
