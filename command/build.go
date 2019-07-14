package command

import (
	"fmt"

	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/debug"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/pkg/errors"
)

type buildCommand struct {
	Filename    string `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	DebugOutput string `long:"debug-output" short:"d" choice:"dot" choice:"json" description:"dump llb to stdout"`
}

func (c *buildCommand) Execute(args []string) (err error) {
	cfg, err := config.ParseFile(c.Filename)
	if err != nil {
		err = errors.Wrapf(err, "failed ot parse config file %s", c.Filename)
		return
	}

	llb, err := frontend.ToLLB(cfg)
	if err != nil {
		err = errors.Wrapf(err, "failed to convert to llb")
		return
	}

	if c.DebugOutput != "" {
		var (
			dotOutput = false
			graph     string
		)

		if c.DebugOutput == "dot" {
			dotOutput = true
		}

		graph, err = debug.LLBToGraph(llb, dotOutput)
		if err != nil {
			err = errors.Wrapf(err, "failed to generate graph")
			return
		}

		fmt.Print(graph)
		return
	}

	return
}
