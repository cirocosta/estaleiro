package frontend

import (
	"github.com/cirocosta/estaleiro/config"
	"github.com/moby/buildkit/client/llb"
)

func ToLLB(cfg *config.Config) (state llb.State, err error) {
	state = llb.Image(cfg.Image.BaseImage.Name)
	return
}
