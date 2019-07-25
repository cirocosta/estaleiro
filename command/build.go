package command

import (
	"context"
	"os"

	"github.com/containerd/console"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type buildCommand struct {
	Address          string            `long:"addr" description:"buildkitd's address"`
	LocalDirectories map[string]string `long:"local" description:"local directory to expose as context (can be specified multiple times)"`
	Filename         string            `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	Variables        map[string]string `long:"var" short:"v" description:"variables to interpolate"`
}

func (c *buildCommand) Execute(args []string) (err error) {
	ctx := context.TODO()

	client, err := bkclient.New(ctx, c.Address, bkclient.WithFailFast())
	if err != nil {
		err = errors.Wrapf(err,
			"failed to construct client for addr %s", c.Address)
		return
	}

	solveOpt := bkclient.SolveOpt{
		LocalDirs: c.LocalDirectories,
	}

	def, _, err := HCLToLLB(c.Filename, c.Variables)
	if err != nil {
		return
	}

	var (
		ch = make(chan *bkclient.SolveStatus)
		eg *errgroup.Group
	)

	eg, ctx = errgroup.WithContext(ctx)

	// initiate the request
	eg.Go(func() (err error) {
		_, err = client.Solve(ctx, def, solveOpt, ch)
		if err != nil {
			err = errors.Wrapf(err,
				"failed while solving")
			return
		}

		return
	})

	// display progress
	eg.Go(func() (err error) {
		var c console.Console

		err = progressui.DisplaySolveStatus(context.TODO(), "", c, os.Stderr, ch)
		if err != nil {
			err = errors.Wrapf(err,
				"failed while displaying status")
			return
		}

		return
	})

	err = eg.Wait()
	return
}
