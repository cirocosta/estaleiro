# a container image builder that (sort of) adheres to the metadata standard for OSL

Hi!

In the past few freedom fridays and flex hours, I've been working on a [proof-of-concept tool][estaleiro] that aims at creating container images in a way that we could more confidently say what is there in the container image, as well as where those bits came from, mostly adhering to the draft [metadata standard][metadata-standard] that the Navcon team shared this week.



```


                                          .----------------------.
      .---------.                         |                      |     .-------------------.
      |  source | === (build process) ==> |   container image    |  +  | bill of materials |
      *---------*           .             |                      |     *-------------------*
                            .             *----------------------*
                            .
                            .
                            ..... `estaleiro`!

                            https://github.com/pivotal/estaleiro

```


What I arrived at is something that allows us go from a sample container image like definition like this:

```hcl
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

  entrypoint = ["/usr/local/bin/estaleiro", "frontend"]
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

to a perfectly valid OCI container image and a bill of materials like this:

```yaml
product_name: cirocosta/estaleiro
protocol: "v0.0.1"
base_image:
    canonical_name: docker.io/library/ubuntu:latest@sha256:c303f19cfe9ee92badbbbd7567bc1ca47789f79303ddcef56f77687d4744cd7a
    os: ubuntu
    version: "18.04"
    codename: bionic
    packages:
      - name: fdisk
        version: 2.31.1-0.4ubuntu3.3
        source_package: util-linux
        architecture: amd64
      - name: libpam-runtime
        version: 1.1.8-3.6ubuntu2.18.04.1
        source_package: pam
        architecture: all
      - name: libncurses5
        version: 6.1-1ubuntu1.18.04
        source_package: ncurses
        architecture: amd64
     # ... more
changeset:
    files:
      - path: /usr/local/bin/estaleiro
        digest: "sha256:5d1f3813d08a162697c1f51bfdb988b8539447aad2e93acfa904741d432ec95b"
        source:
            git:
                repository_uri: https://github.com/cirocosta/estaleiro
                ref: 49b6936ef130793fff677038891e517718c2baf8
    packages:
      - name: libssl1.1
        version: 1.1.1-1ubuntu2.1~18.04.4
        source_package: openssl
        architecture: amd64
        location:
            uri: http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1-1ubuntu2.1~18.04.4_amd64.deb
            digest: md5:41f3ea2b9f5b419550f165975b941f81
            name: libssl1.1_1.1.1-1ubuntu2.1~18.04.4_amd64.deb
        sources:
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/openssl_1.1.1-1ubuntu2.1~18.04.4.dsc
            digest: md5:88218150efac41c72aaf0025cd4d481800e0871e5ea045d25c5b10b09f7b0a88
            name: openssl_1.1.1-1ubuntu2.1~18.04.4.dsc
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/openssl_1.1.1.orig.tar.gz
            digest: md5:2836875a0f89c03d0fdf483941512613a50cfb421d6fd94b9f41d7279d586a3d
            name: openssl_1.1.1.orig.tar.gz
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/openssl_1.1.1.orig.tar.gz.asc
            digest: md5:f3296150114069ea73a72eafbfdcbb295b770e7cbf3266f9590f3d0932498b3e
            name: openssl_1.1.1.orig.tar.gz.asc
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/openssl_1.1.1-1ubuntu2.1~18.04.4.debian.tar.xz
            digest: md5:a373c2612817f3ae929d01ddb9175a6f9ab0ac28a08afe93c88df27fadcc7500
            name: openssl_1.1.1-1ubuntu2.1~18.04.4.debian.tar.xz
      - name: ca-certificates
        version: "20180409"
        source_package: ""
        architecture: all
        location:
            uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409_all.deb
            digest: md5:eae40792673dcb994af86284d0a01f36
            name: ca-certificates_20180409_all.deb
        sources:
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409.dsc
            digest: md5:cd1f6540d0dab28f897e0e0cb2191130cdbf897f8ce3f52c8e483b2ed1555d30
            name: ca-certificates_20180409.dsc
          - uri: http://archive.ubuntu.com/ubuntu/pool/main/c/ca-certificates/ca-certificates_20180409.tar.xz
            digest: md5:7af6f5bfc619fd29cbf0258c1d95107c38ce840ad6274e343e1e0d971fc72b51
            name: ca-certificates_20180409.tar.xz
      # ... more
```


**This is still an experiment** that I'm sharing with folks around Pivotal to gather some feedback and see if I can get us thinking about other possibilities around the way that we build container images that might not fit the models that we have right now.

Please give me feedback!


## context

As an outsider to the work that is being put into this area by other teams (I'm part of Concourse for PCF), I focused on trying to see if there was a way of making our lives in Concourse by the time that we get to ship a container image to our customers - some of our images deviate from the traditional "build go and copy the binary".

So I started experimenting - tring to come up with ways of making whatever we put into that `(build process)` box there in the diagram above produce some kind of artifact that would give us visibility into what went on during that build, while still forcing us to keep track of where things came from.


```

   
    declaration of the image we -.
    want to produce              |
                                 |
            +                    |                               container image
                                 |     ===> build process ===>          +
                                 |                                bill of materials
         context                 |
                                 |
    (additional files to add) ---*
   

```

It turns out that during the movement of splitting the Docker engine into multiple components, one of them (the builder) got completely rewritten in a way that allows us to build on top of it, leveraging the primitives of building container images, while being able to define whatever syntax (and semantics) we want - this is what `estaleiro` leverages.


ps.: I talked about it in the Toronto office last week (see the talk in a readable format here: [talk][talk])



## how it works

`estaleiro` implements a [buildkit][buildkit] frontend that interprets that initial declaration of how a container image should look like and creates the corresponding intermediary representation (IR) that `buildkitd` (the piece that goes from an IR to a series of steps to build the final image) consumes.


```

      estaleiro.hcl  
         ===>  estaleiro frontend 
            ===> LLB  (IR)
                ==> buildkitd 
                    ==> image + bill of materials

```

As we define what each step means (e.g., adding packages, adding files ...), that means that we can do whatever we want with those bits - computing the checksums, annotating what came in, enforcing certain rules, etc.  


```

    >>>  "I want to add 3 packages!"


                   "Sure, I'll see if I can add them for you"  <<<

                  1. validates a bunch of stuff before adding
                  2. writes down what those packages and their dependencies are

```

At the moment, `estaleiro` has three ways of adding content to the final image:

1. apt repositories
2. local files from tarballs
3. files from the result of building Dockerfiles

Naturally, this can be extended / modified - the whole focus that I had was making Concourse images be buildable with it, thus the bias towards those three (although I don't see that list growing much more).

Under the hood, a definition like that one that I showed in the beginning, gets translated into a directed acyclic graph that represents the steps that `buildkit` goes through that from a very high level looks like this two tree intertwined:


```

      FINAL FS                       BILL OF MATERIALS FS
  (layers that get into the            (a "side-tree" that just exists
    container image)                    outside the final container image)



      grab base image

                                      collect base image info (os-release)
                                      collect base image packages info


      install packages
                                      collect packages info + deb packages source


      add files from tarballs

                                      write down the procedence of those + metadata


      add files from images built
                          
                                       write down the procedent of those + metadata



                           final layer


```

![](graph.png)



## the developer experience

For a developer, very few aspects of his/her daily work change - after an `estaleiro.hcl` file gets created, regular `docker build`s can be performed against it - since newer version of Docker, it can leverage any Buildkit frontend implementation to create that IR that tells Docker how to build your container image.

For instance, if you want to try this out, head to https://github.com/pivotal/estaleiro and build the `estaleiro` image itself:


```console
# tell Docker to leverage buildkit under the hood
#
$ export DOCKER_BUILDKIT=1


# build using the `estaleiro` frontend
#
$ docker build \
  --file ./estaleiro.hcl \
  --tag sample \
  --build-arg estaleiro-commit=$(git rev-parse HEAD)\
  .
```

This is possible by leveraging the `# syntax = cirocosta/estaleiro` in the first line of the `estaleiro.hcl` file - `docker` parses that line and then fetches the container image under `cirocosta/estaleiro`, which then uses to perform the interpretation of the `estaleiro.hcl` file.  

## the CI experience

As regular `docker` can be used in your own machine, not much change is needed for a CI that can leverage `docker` inside their tasks.  

It's also possible to use buildkit directly to perform the builds, thus, not changing the usual CI flow.


## what's next?

I'm not sure yet!

This is just a side project that I'm excited about, working on it during flex hours and freedom fridays.

I plan to:

- continue sharing what I learn with it
- eventually arrive at a CF RFC dependending on how the feedback goes
- ?

Thank you!


[buildkit]: https://github.com/moby/buildkit
[estaleiro]: https://github.com/pivotal/estaleiro
[talk]: https://github.com/cirocosta/estaleiro/blob/812821d513103f1157956a925a696267374add90/docs/talk/README.md
[metadata-standard]: https://docs.google.com/document/d/1o5zVpEva8EBbMmdREUYkJpcCrCuNqA-lbpFI9ri7n88/edit

