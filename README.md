
# estaleiro [istaˈlejru]

<br />
<br />

<img align="left" width="384" height="256" src="https://github.com/cirocosta/estaleiro/raw/master/.github/shipyard.jpg" />

<br />
<br />


> *masculine noun - shipyard*

`estaleiro` allows you to ship container images with confidence - a declarative
approach to dealing with the last mile in building container images, so you can
have more control (through transparency) over what you ship.

<br />
<br />
<br />

**HIGHLY EXPERIMENTAL - DO NOT USE THIS**

<br />

**Table of Contents**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [problem set](#problem-set)
- [estaleiro](#estaleiro)
- [when will I be able to use this?](#when-will-i-be-able-to-use-this)
- [license](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->



### problem set



Keeping track of what has been added to a container image that one is about to
ship if hard.

With the versatility of Dockerfiles, it's quite easy to shoot itself in the foot
when it comes to either installing dependencies that one couldn't even consume,
or that wouldn't be wise.

While it's great to talk about best practices, it's hard to enforce them.



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

# the final image to produce
#
image "cirocosta/estaleiro" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }
  
  apt {
    package "ca-certificates" {}
  }

  file "/usr/local/bin/estaleiro" {
    from_step "build" {
      path = "/bin/estaleiro"
    }
  }
}


# performs the build of `estaleiro`.
#
step "build" {
  dockerfile = "./Dockerfile"
  target     = "build"

  source_file "/bin/estaleiro" {
    vcs "git" {
      ref        = "${estaleiro-commit}"
      repository = "https://github.com/cirocosta/estaleiro"
    }
  }
}
```

Having those pieces in, `estaleiro` creates the intermediary representation to
be used by `buildkitd` to build the final container image that starts from
`ubuntu:bionic`, has the `ca-certificates` package installed, and the file that
the Dockerfile built - all while keeping track of their versions and sources
along the way in the form of a bill of materials:


```yaml
base_image:
  name: docker.io/library/ubuntu
  digest: sha256:c303f19cfe9ee92badbbbd7567bc1ca47789f79303ddcef56f77687d4744cd7a
  packages:
    - name: fdisk
      version: 2.31.1-0.4ubuntu3.3
      source_package: util-linux
      architecture: amd64
    - name: libpam-runtime
      version: 1.1.8-3.6ubuntu2.18.04.1
      source_package: pam
      architecture: all
    # ...

changeset:
  files:
    - name: "/usr/local/bin/seataleiro"
      digest: "sha256:89f687d4744cd779303ddc7ef56f77c303f19cfe9ee92badbbbd7567bc1ca47a"
      source:
        - url: https://github.com/cirocosta/estaleiro
          type: git
          ref: 6a4d0b73673a1863a62b7ac6cbde4ae7597c56d7
      from_step:
        name: "build"
        dockerfile_digest: "sha256:9303ddc7ef56f77c303f19cfe9ee92badbbbd7567bc189f687d4744cd77ca47a"
  packages:
    - name: ca-certificates
      version: "20180409"
      source_package: ""
      architecture: all
      location:
          uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409_all.deb
          name: ca-certificates_20180409_all.deb
          size: "150932"
          md5sum: eae40792673dcb994af86284d0a01f36
      source:
        - uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409.dsc
          name: ca-certificates_20180409.dsc
          size: "1420"
          md5sum: cd1f6540d0dab28f897e0e0cb2191130cdbf897f8ce3f52c8e483b2ed1555d30
        - uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409.tar.xz
          name: ca-certificates_20180409.tar.xz
          size: "246908"
          md5sum: 7af6f5bfc619fd29cbf0258c1d95107c38ce840ad6274e343e1e0d971fc72b51
    # and all of its dependencies too ...
```


### how to use

**THIS IS STILL HIGHLY EXPERIMENTAL**

All that you need is:

- Docker 18.09+

Having an `estaleiro` file (like the `estaleiro.hcl` that you find in this
repo), direct `docker build` to it via a regular `--file` (`-f`), having
`DOCKER_BUILDKIT=1` set as the environent variable:

```console
# create an `estaleiro.hcl` file.
# 
# note.: the first line (with syntax ...) is important - it's
#        what tells the docker engine to fetch our implementation
#        of `estaleiro`, responsible for creating the build
#        definition.
#
$ echo "# syntax=cirocosta/estaleiro
image "cirocosta/sample" {
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }
}
" > ./estaleiro.hcl


# instruct `docker` to build our image
#
$ docker build -t test -f ./estaleiro.hcl
[+] Building 9.4s (4/4) FINISHED


# retrieve the bill of materials from the filesystem
#
$ docker create --name tmp
$ docker cp tmp:/bom/merged.yml ./bom.yml
```


### references


- [buildpacks - RFC for additional metadata](https://github.com/buildpack/rfcs/pull/9)
- [doc - Container Image metadata standard for OSL process](https://docs.google.com/document/d/1o5zVpEva8EBbMmdREUYkJpcCrCuNqA-lbpFI9ri7n88)
- [doc - App Image Metadata, a Taxonomy](https://docs.google.com/document/d/1ITAxZKZmF802PHmXmsEN_FqK1VMGol8N_TLHECUXDNU)


### license

See [./LICENSE](./LICENSE).

