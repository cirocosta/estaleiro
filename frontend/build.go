package frontend

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

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

const (
	localName            = "dockerfile"
	keyTarget            = "target"
	keyFilename          = "filename"
	keyCacheFrom         = "cache-from"
	defaultConfigName    = "estaleiro.hcl"
	dockerignoreFilename = ".dockerignore"
	buildArgPrefix       = "build-arg:"
	labelPrefix          = "label:"
	keyNoCache           = "no-cache"
	keyImageResolveMode  = "image-resolve-mode"
)

func Build(ctx context.Context, client gateway.Client) (res *gateway.Result, err error) {
	state, img, err := prepareLLBFromClient(ctx, client)
	if err != nil {
		err = errors.Wrapf(err, "failed to build for linux amd64")
		return
	}

	res, err = invokeBuild(ctx, client, state, img)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to make build call")
		return
	}

	return
}

// prepareLLBFromClient retrieves the options provided through the client call
// and assemblies the LLB to build it.
//
func prepareLLBFromClient(
	ctx context.Context, c gateway.Client,
) (
	state llb.State, img ocispec.Image, err error,
) {
	cfg, err := readConfigFromClient(ctx, c)
	if err != nil {
		err = errors.Wrapf(err,
			"failed reading config from client")
		return
	}

	dockerfileMapping, err := gatherDockerfiles(ctx, c, cfg)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to read dockerfiles from context")
		return
	}

	state, img, err = ToLLB(ctx, cfg, dockerfileMapping)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to generate llb from file")
		return
	}

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

	// read a file from this part

	materials, err := ref.ReadFile(ctx, gateway.ReadRequest{
		Filename: "/bom/merged.yml",
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to read merged bom")
		return
	}

	img.Config.Labels = map[string]string{
		"estaleiro.bom": base64.StdEncoding.EncodeToString(materials),
	}

	config, err := json.Marshal(img)
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal image config")
		return
	}

	ioutil.WriteFile("/tmp/config", config, 0644)

	k := platforms.Format(platforms.DefaultSpec())

	res.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, k), config)
	res.SetRef(ref)

	return
}

func gatherDockerfiles(ctx context.Context, c gateway.Client, cfg *config.Config) (mapping map[string][]byte, err error) {
	mapping = make(map[string][]byte)

	for _, step := range cfg.Steps {
		_, found := mapping[step.Dockerfile]
		if found {
			continue
		}

		mapping[step.Dockerfile], err = readDockerfile(ctx, c, step.Dockerfile)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to read dockerfile from local context")
			return
		}
	}

	return
}

func readDockerfile(ctx context.Context, c gateway.Client, name string) (content []byte, err error) {
	src := llb.Local(localName,
		llb.IncludePatterns([]string{name}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint("dockerfile"),
		llb.WithCustomName("[estaleiro] load dockerfile "+name),
	)

	def, err := src.Marshal()
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal local source")
		return
	}

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

	content, err = ref.ReadFile(ctx, gateway.ReadRequest{
		Filename: name,
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to read dockerfile")
		return
	}

	return
}

func readConfigFromClient(ctx context.Context, c gateway.Client) (cfg *config.Config, err error) {
	opts := c.BuildOpts().Opts

	filename := opts[keyFilename]
	if filename == "" {
		filename = defaultConfigName
	}

	src := llb.Local(localName,
		llb.IncludePatterns([]string{filename}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint(defaultConfigName),
		llb.WithCustomName("[estaleiro] load estaleiro file "+filename),
	)

	def, err := src.Marshal()
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal local source")
		return
	}

	res, err := c.Solve(ctx, gateway.SolveRequest{
		Definition: def.ToPB(),
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to resolve request for local source")
		return
	}

	ref, err := res.SingleRef()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve single ref from resolution")
		return
	}

	var hclBytes []byte
	hclBytes, err = ref.ReadFile(ctx, gateway.ReadRequest{
		Filename: filename,
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to read config")
		return
	}

	vars := filter(opts, buildArgPrefix)

	cfg, err = config.Parse(hclBytes, filename, vars)
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
