
# estaleiro [istaˈlejru]

<br />
<br />

<img align="left" width="384" height="256" src="https://github.com/cirocosta/estaleiro/raw/master/.github/shipyard.jpg" />

<br />
<br />


> *masculine noun - shipyard*

`estaleiro` allows you to ship container images with confidence - a declarative
approach to dealing with the last mile in building container images, so you can
have more control over what you ship.

<br />
<br />
<br />
<br />

**Table of Contents**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [problem set](#problem-set)
- [estaleiro](#estaleiro)
- [developing](#developing)
- [license](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->



### problem set

> consistently enforcing a guideline for baking container images is hard

`Dockerfile`s are great - they allow us to do anything!

With that power, problems arise when an entire organization starts adopting it:

- no control over base images
- some drop user privileges, others don't
- artifacts can end up in them coming from who knows where
  - with which versions?
  - having the source code from where?



### estaleiro

`estaleiro` leverages [`buildkit`](https://github.com/moby/buildkit) as a way of
implementing a convention of how the last stage in building a container image
(i.e., gathering binaries built in previous steps), putting guard-rails where
needed, and enforcing a set of rules where needed.

Here's an example of how that looks like in practice:

1. bring your Dockerfile that you've already been using to build your binary

```Dockerfile
FROM golang AS base

	ENV CGO_ENABLED=0
	RUN apt update && apt install -y git

	ADD . /src
	WORKDIR /src

	RUN go mod download


FROM base AS build

	RUN go build \
		-tags netgo -v -a \
		-o /usr/local/bin/estaleiro \
		-ldflags "-X main.version=$(cat ./VERSION) -extldflags \"-static\""
```

2. bring a `estaleiro` file that describes how to package that binary produced

```hcl
# syntax = cirocosta/estaleiro-frontend

step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"
}

image "cirocosta/estaleiro" {

  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

  file "/usr/local/bin/estaleiro" {
    from_step "build" {
      path = "/usr/local/bin/estaleiro"
    }
  }
}
```


### tests
At the moment, there are no integration tests.

Unit tests can be run with the standard `go test`:

```console
$ go test -v ./...
?   	github.com/cirocosta/estaleiro	[no test files]
?   	github.com/cirocosta/estaleiro/command	[no test files]
=== RUN   TestConfig
Running Suite: Config Suite
...
```


### developing it locally

1. start `buildkitd` inside a container (that can be reached from the host).

```console
$ make run-buildkitd
```


2. with `buildkit` running, set the environment variable `BUILDKIT_HOST` that
   allows `estaleiro` to target `buildkitd`.

```console
$ export BUILDKIT_HOST=tcp://0.0.0.0:1234
```

Now, you can run `estaleiro` using the standard `buildctl` CLI to interact with
`buildkitd`.

For instance, to build an image that ships `estaleiro`:

```console
./estaleiro build -f ./estaleiro.hcl | buildctl build --local context=.
```


### license

See [./LICENSE](./LICENSE).

