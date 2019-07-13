package main

import (
	"fmt"
	"os"

	"github.com/cirocosta/estaleiro/config"
	"github.com/jessevdk/go-flags"
)

var cli struct {
	filename string `long:"filename" short:"f" required:"true" description:"file containing image definition"`
}

func main() {
	parser := flags.NewParser(&cli, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing cli arguments: %v\n", err)
		return
	}

	_, err = config.ParseFile(cli.filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse configuration file: %v\n", err)
		return
	}

	return
}
