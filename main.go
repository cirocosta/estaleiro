package main

import (
	"fmt"
	"os"

	"github.com/cirocosta/estaleiro/config"
	"github.com/cirocosta/estaleiro/debug"
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/jessevdk/go-flags"
)

var cli struct {
	Filename    string `long:"filename" short:"f" required:"true" description:"file containing image definition"`
	DebugOutput string `long:"debug-output" choice:"dot" choice:"json" description:"dump llb to stdout"`
}

func main() {
	parser := flags.NewParser(&cli, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing cli arguments: %v\n", err)
		return
	}

	cfg, err := config.ParseFile(cli.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse configuration file: %v\n", err)
		return
	}

	llb, err := frontend.ToLLB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert to llb: %v\n", err)
		return
	}

	if cli.DebugOutput != "" {
		var dotOutput = false

		if cli.DebugOutput == "dot" {
			dotOutput = true
		}

		graph, err := debug.LLBToGraph(llb, dotOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate graph: %v\n", err)
			return
		}

		fmt.Print(graph)
	}

	return
}
