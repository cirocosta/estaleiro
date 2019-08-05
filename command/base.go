package command

import (
	"fmt"

	"github.com/cirocosta/estaleiro/dpkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type baseCommand struct {
	Output string `long:"output"      required:"true" description:"where to write the bill of materials to ('-' for stdout)"`
}

type OsReleaseV1 struct {
	Kind string         `yaml:"kind"`
	Data dpkg.OsRelease `yaml:"data"`
}

func NewOsReleaseV1(info dpkg.OsRelease) OsReleaseV1 {
	return OsReleaseV1{
		Kind: "osrelease/v1",
		Data: info,
	}
}

func (c *baseCommand) Execute(args []string) (err error) {
	writer, err := writer(c.Output)
	if err != nil {
		return
	}

	info, err := dpkg.GatherOsRelease()
	if err != nil {
		return
	}

	b, err := yaml.Marshal(NewOsReleaseV1(info))
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(writer, "%s", string(b))
	if err != nil {
		err = errors.Wrapf(err,
			"failed writing bill of materials to %s", c.Output)
		return
	}

	return
}
