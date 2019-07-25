package command

import (
	"context"

	bkclient "github.com/moby/buildkit/client"
	"github.com/pkg/errors"
)

type buildCommand struct {
	Address          string            `long:"addr" description:"buildkitd's address"`
	LocalDirectories map[string]string `long:"local" description:"local directory to expose as context (can be specified multiple times)"`
	Filename         string            `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	Variables        map[string]string `long:"var" short:"v" description:"variables to interpolate"`
}

func (c *buildCommand) Execute(args []string) (err error) {
	client, err := bkclient.New(context.TODO(), c.Address, bkclient.WithFailFast())
	if err != nil {
		err = errors.Wrapf(err,
			"failed to construct client for addr %s", c.Address)
		return
	}

	solveOpt := bkclient.SolveOpt{
		LocalDirs: c.LocalDirectories,
	}

	def, bom, err := HCLToLLB(c.Filename, c.Variables)
	if err != nil {
		return
	}

	return
}
