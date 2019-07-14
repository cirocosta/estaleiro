package command

import (
	"github.com/cirocosta/estaleiro/frontend"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
)

type frontendCommand struct{}

func (c *frontendCommand) Execute(args []string) (err error) {
	err = grpcclient.RunFromEnvironment(appcontext.Context(), frontend.Build)

	if err != nil {
		err = errors.Wrapf(err, "failed to run builder")
		return
	}

	return
}
