package frontend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cirocosta/estaleiro/config"
	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

var (
	linuxAMD64 = ocispec.Platform{OS: "linux", Architecture: "amd64"}
)

// prepareLLBFromClient retrieves the options provided through the client call
// and assemblies the LLB to build it.
//
func prepareLLBFromClient(
	ctx context.Context, client gateway.Client,
) (
	state llb.State, img ocispec.Image, err error,
) {
	cfg, err := readConfigFromClient(ctx, client)
	if err != nil {
		return
	}

	state, err = ToLLB(&cfg)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to generate llb from file")
		return
	}

	// TODO this should come from `2llb`
	img = ocispec.Image{
		Architecture: "amd64",
		OS:           "linux",
		Config: ocispec.ImageConfig{
			Cmd: []string{"/bin/false"},
		},
	}

	return
}

func Build(ctx context.Context, client gateway.Client) (res *gateway.Result, err error) {
	state, img, err := prepareLLBFromClient(ctx, client)
	if err != nil {
		err = errors.Wrapf(err, "failed to build for linux amd64")
		return
	}

	res, err = invokeBuild(ctx, client, state, img)
	if err != nil {
		err = errors.Wrapf(err,
			"faoled to make build call")
		return
	}

	return
}

func readConfigFromClient(ctx context.Context, client gateway.Client) (cfg config.Config, err error) {
	return
}

func invokeBuild(
	ctx context.Context, client gateway.Client, state llb.State, img ocispec.Image,
) (
	res *gateway.Result, err error,
) {
	def, err := state.Marshal()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to marshal llb state into protobuf definition")
		return
	}

	res, err = client.Solve(ctx, gateway.SolveRequest{
		Definition: def.ToPB(),
	})
	if err != nil {
		err = errors.Wrapf(err,
			"failed performing solve request")
		return
	}

	ref, err := res.SingleRef()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve single ref")
		return
	}

	config, err := json.Marshal(img)
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal image config")
		return
	}

	k := platforms.Format(platforms.DefaultSpec())

	res.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, k), config)
	res.SetRef(ref)

	return
}
