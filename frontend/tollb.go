package frontend

import (
	"bytes"

	"github.com/cirocosta/estaleiro/config"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

func ToLLB(cfg *config.Config) (state llb.State, err error) {
	state = llb.Image(cfg.Image.BaseImage.Name)
	return
}

func LLBToGraph(state llb.State) (res string, err error) {
	var (
		definition *llb.Definition
		buffer     bytes.Buffer
	)

	definition, err = state.Marshal()
	if err != nil {
		err = errors.Wrap(err, "marshaling llb state")
		return
	}

	err = llb.WriteTo(definition, &buffer)
	if err != nil {
		err = errors.Wrap(err, "failed writing llb definition to buffer")
		return
	}

	res = buffer.String()

	return
}
