package command

import (
	"os"

	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

type buildCommand struct {
	Filename string `long:"filename" short:"f" required:"true" description:"file containing image definition"`
}

func (c *buildCommand) Execute(args []string) (err error) {
	cfg, err := config.ParseFile(c.Filename)
	if err != nil {
		err = errors.Wrapf(err, "failed ot parse config file %s", c.Filename)
		return
	}

	state, err := frontend.ToLLB(cfg)
	if err != nil {
		err = errors.Wrapf(err, "failed to convert to llb")
		return
	}

	definition, err := state.Marshal()
	if err != nil {
		err = errors.Wrap(err, "marshaling llb state")
		return
	}

	llb.WriteTo(definition, os.Stdout)

	return
}
