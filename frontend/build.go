package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cirocosta/estaleiro/config"
	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

var (
	linuxAMD64 = ocispec.Platform{OS: "linux", Architecture: "amd64"}
)

const (
	LocalNameContext      = "context"
	LocalNameDockerfile   = "dockerfile"
	keyTarget             = "target"
	keyFilename           = "filename"
	keyCacheFrom          = "cache-from"
	defaultDockerfileName = "estaleiro.hcl"
	dockerignoreFilename  = ".dockerignore"
	buildArgPrefix        = "build-arg:"
	labelPrefix           = "label:"
	keyNoCache            = "no-cache"
	keyTargetPlatform     = "platform"
	keyImageResolveMode   = "image-resolve-mode"
	keyGlobalAddHosts     = "add-hosts"
	keyForceNetwork       = "force-network-mode"
	keyOverrideCopyImage  = "override-copy-image" // remove after CopyOp implemented
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

	state, img, _, err = ToLLB(cfg)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to generate llb from file")
		return
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

func readConfigFromClient(ctx context.Context, c gateway.Client) (cfg *config.Config, err error) {
	opts := c.BuildOpts().Opts

	filename := opts[keyFilename]
	if filename == "" {
		filename = defaultDockerfileName
	}

	name := "load Estaleiro file"
	if filename != defaultDockerfileName {
		name += " from " + filename
	}

	src := llb.Local(LocalNameDockerfile,
		llb.IncludePatterns([]string{filename}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint(defaultDockerfileName),
		dockerfile2llb.WithInternalName(name),
	)

	def, err := src.Marshal()
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal local source")
		return
	}

	var dtDockerfile []byte
	res, err := c.Solve(ctx, gateway.SolveRequest{
		Definition: def.ToPB(),
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to resolve dockerfile")
		return
	}

	ref, err := res.SingleRef()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve single ref from resolution")
		return
	}

	dtDockerfile, err = ref.ReadFile(ctx, gateway.ReadRequest{
		Filename: filename,
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to read dockerfile")
		return
	}

	vars := filter(opts, buildArgPrefix)

	cfg, err = config.Parse(dtDockerfile, filename, vars)
	if err != nil {
		err = errors.Wrapf(err, "failed parsing config")
		return
	}

	return
}

func filter(opt map[string]string, key string) map[string]string {
	m := map[string]string{}
	for k, v := range opt {
		if strings.HasPrefix(k, key) {
			m[strings.TrimPrefix(k, key)] = v
		}
	}
	return m
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
