package main

import (
	"fmt"
	"os"

	"github.com/cirocosta/estaleiro/command"
	"github.com/jessevdk/go-flags"
)

var cli struct {
}

func main() {
	parser := flags.NewParser(&command.Estaleiro, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing cli arguments: %v\n", err)
		return
	}

}
