package command

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
)

type retrieveBomCommand struct {
	Image         string `long:"image"`
	DockerTarball string `long:"docker-tarball"`
}

func (c *retrieveBomCommand) Execute(args []string) (err error) {
	var image v1.Image

	switch {
	case c.Image != "":
		image, err = resolveFromRegistry(c.Image)
	case c.DockerTarball != "":
		image, err = tarball.ImageFromPath(c.DockerTarball, nil)
		if err != nil {
			err = errors.Wrapf(err, "could not load image from path '%s'", c.DockerTarball)
			return
		}

	default:
		err = errors.Errorf("either `image` or `docker-tarball` must be specified")
	}
	if err != nil {
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

func resolveFromRegistry(imageName string) (image v1.Image, err error) {
	n, err := name.ParseReference(imageName, name.WeakValidation, name.Insecure)
	if err != nil {
		err = errors.Wrapf(err, "could not resolve repository/tag reference '%s'", imageName)
		return
	}

	image, err = remote.Image(n)
	if err != nil {
		err = errors.Wrapf(err, "failed to get remote image %s", image)
		return
	}

	return
}
