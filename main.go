package main

import (
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/cirocosta/statuscheck/commands"
	"github.com/jessevdk/go-flags"
)

func main() {
	logger := lager.NewLogger("my-app")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	parser := flags.NewParser(&commands.StatusCheck, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		logger.Error("parsing", err)
		os.Exit(1)
	}

	return
}
