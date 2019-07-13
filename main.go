package main

import (
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
)

var (
	config struct{}
)

func main() {
	logger := lager.NewLogger("estaleiro")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	parser := flags.NewParser(&config, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		logger.Error("parsing", err)
		os.Exit(1)
	}

	return
}
