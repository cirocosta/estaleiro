package command

import (
	"fmt"
	"io"
	"os"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type collectCommand struct {
	Input  string `long:"input"  required:"true" description:"path to dpkg status file ('-' for stdin)"`
	Output string `long:"output" required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

type Packages []dpkg.Package

func (p Packages) ToYAML() (res []byte) {
	var err error

	res, err = yaml.Marshal(&p)
	if err != nil {
		panic(err)
	}

	return
}

func (c *collectCommand) Execute(args []string) (err error) {
	var packages []dpkg.Package

	input, err := reader(c.Input)
	if err != nil {
		return
	}

	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	scanner := dpkg.NewScanner(input)

	packages, err = scanner.ScanAll()
	if err != nil {
		err = errors.Wrapf(err,
			"failed scanning packages from dpkg status file %s",
			c.Input)
		return
	}

	fmt.Fprintf(writer, "%+v\n", string(Packages(packages).ToYAML()))

	return
}

func writer(value string) (w io.Writer, err error) {
	var file *os.File

	if value == "-" {
		w = os.Stdout
		return
	}

	file, err = os.Create(value)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create file %s", value)
		return
	}

	w = file
	return
}

func reader(value string) (r io.Reader, err error) {
	var file *os.File

	if value == "-" {
		r = os.Stdin
		return
	}

	file, err = os.Open(value)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to open dpkg status file at %s", value)
		return
	}

	r = file
	return
}
