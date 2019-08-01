module github.com/cirocosta/estaleiro

go 1.12

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0-0.20190426060238-3a3f0aac8819
	github.com/docker/distribution v2.7.1+incompatible
	github.com/fatih/color v1.7.0
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20190725010614-0c3fe388e450
	github.com/jessevdk/go-flags v1.4.0
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/moby/buildkit v0.5.1
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/zclconf/go-cty v1.1.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/yaml.v3 v3.0.0-20190709130402-674ba3eaed22
)

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
