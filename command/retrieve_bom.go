package command

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
)

type retrieveBomCommand struct {
	Image string `long:"image" required:"true"`
}

func (c *retrieveBomCommand) Execute(args []string) (err error) {
	n, err := name.ParseReference(c.Image, name.WeakValidation, name.Insecure)
	if err != nil {
		err = errors.Wrapf(err, "could not resolve repository/tag reference '%s'", c.Image)
		return
	}

	image, err := remote.Image(n)
	if err != nil {
		err = errors.Wrapf(err, "failed to get remote image %s", c.Image)
		return
	}

	configFile, err := image.ConfigFile()
	if err != nil {
		err = errors.Wrapf(err, "failed to retrieve image's config file")
		return
	}

	labels := configFile.Config.Labels

	bom, found := labels["estaleiro.bom"]
	if !found {
		err = errors.Errorf("`estaleiro.bom` label not found")
		return
	}

	decodedBom, err := base64.StdEncoding.DecodeString(bom)
	if err != nil {
		err = errors.Wrapf(err,
			"failed decoding b64 representation of the `bom`")
		return
	}

	fmt.Printf("%s", string(decodedBom))
	return
}
