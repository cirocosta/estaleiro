<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [rethinking the way container images are built](#rethinking-the-way-container-images-are-built)
  - [the Concourse container image](#the-concourse-container-image)
  - [what's hard about dockerfiles](#whats-hard-about-dockerfiles)
    - [1. tracking what was added](#1-tracking-what-was-added)
    - [2. ensuring consistenty](#2-ensuring-consistenty)
  - [the bare minimum of building container images](#the-bare-minimum-of-building-container-images)
  - [a primitive container image builder](#a-primitive-container-image-builder)
  - [compilers](#compilers)
  - [buildkit](#buildkit)
  - [a minimally viable frontend for our container images](#a-minimally-viable-frontend-for-our-container-images)
  - [installing packages](#installing-packages)
  - [container image as an artifact](#container-image-as-an-artifact)
  - [what enabled this](#what-enabled-this)
  - [from syntax to container image](#from-syntax-to-container-image)
    - [snapshots](#snapshots)
  - [concourse images as an example](#concourse-images-as-an-example)
  - [references](#references)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# rethinking the way container images are built

Hey,

We at the Concourse for PCF team have been maintaining for quite a while a Helm
chart for running Concourse on top of Kubernetes, and, wanting to ship that to
customers, showed that there's quite a good deal of uncertainty involved.

```

  OSS USERS

    `helm/charts/stable/Concourse` ---> powers `hush-house`!

        :D


  CUSTOMERS


        I'd like that too!


  WE

        we're still figuring it out

        ¯\_(ツ)_/¯


```


Although we're fairly used to the traditional ways of getting the Concourse code
in a state where it can be distributed as a BOSH release artifact on PivNet,
we're very new to doing so for customers who would like to have it on their PKS
installations - differently from a BOSH release, now there are container images
and some wrapping around the formation of those Kubernetes objects.

If we consider that there are several steps to go from "we have a container
image" and want to get that to customer hands, there's clearly some steps to go
through. Thinking of it as a Makefile, that'd look like:


```Makefile
customer_artifact: download_from_pivnet


download_from_pivnet: artifact_with_osl_file


artifact_with_osl_file: container_image_that_is_scannable


container_image_that_is_scannable:
    echo "¯\_(ツ)_/¯"
```

That is,


1. **to get that to customers**, *it needs to be on PivNet,*
> so that we can attest what has been distributed to who

2. **to have it on PivNet**, *it needs an OSL file*
> so that we can disitribute the copyright and prove that we don't have any
> licenses that would hurt our customers

3. **to get the OSL file**, *it needs to let norsk know where to scan source code*
> so that it can know what are those licenses, and gather copyright info

4. **to let Norsk know where to scan**, *it needs to `¯\_(ツ)_/¯`*


Thus, with a single goal in mind - shipping Concourse container images to
customers running Kubernetes -, we can see that if at the very bottom of it -
creating a container image that is scannable -, there's friction, the whole
process can get delayed, impacting all of the rest of the process of getting the
great features that the team developed to the hands of our customers.


To make things more concrete, let's look at one of the container images that we
need to scan to ship Concourse.


## the Concourse container image

Although Concourse has more then 10 container images that we publish on
DockerHub as of today that are consumed by OSS folks, we can look at the case of
`concourse/concourse` - the image that provides the Concourse binaries with all
bateries included:


```dockerfile
# defining the base image that will produce the final concourse/concourse
# image - this adheres to pivotal practice of using `ubuntu:bionic` images.
#
FROM ubuntu:bionic AS ubuntu


FROM ubuntu AS assets

  # retrieve the concourse tarball that contains all of the concourse
  # binaries and other necessary artifacts.
  #
  COPY ./linux-rc/*.tgz /tmp
  RUN tar xzf /tmp/*tgz -C /usr/local


FROM ubuntu

  # some environment variables (there are more)
  ENV CONCOURSE_SESSION_SIGNING_KEY     /concourse-keys/session_signing_key
  ENV CONCOURSE_TSA_AUTHORIZED_KEYS     /concourse-keys/authorized_worker_keys



  # volume for non-aufs/etc. mount for baggageclaim's driver
  VOLUME /worker-state


  # packages needed at runtime
  RUN apt update && apt install -y \
      btrfs-tools \
      ca-certificates \
      dumb-init \
      iproute2 \
      file


  # retrieve the bits that were extracted in the previous step.
  #
  COPY --from=assets /usr/local/concourse /usr/local/concourse

  STOPSIGNAL SIGUSR2
  ENTRYPOINT ["dumb-init", "/usr/local/concourse/bin/concourse"]
```



## what's hard about dockerfiles

While dockerfiles are great for the versatility that they provide, some
challanges start arising for Rachel.

### 1. tracking what was added

Although the above Dockerfile seems very straightforward, it presents some
difficulties when it comes to preciselly telling what's there:

- what was the SHA of that tarball that brought contents?
- how can I tie that tarball to source code?
- what are the versions of those packages?
- can I get the source of those packages?
- what debian repository brought those packages?

It turns out that due to the extreme versatility of a Dockerfile, it's very
tough for a company to get that information precisely.

```

     concourse/concourse
    .-------------------.
    | 
    |   btrfs-progs                           version=?
    |   btrfs-tools                           sha=?
    |   ca-certificates                       repository?
    |   /usr/local/concourse/bin/concourse    tarball_sha=?
    |   /usr/local/concourse/bin/gdn          source=?
    |   ...


```

### 2. ensuring consistenty

While it's known that some best practices exist around creating container
images, Dockerfiles lack ways of enforcing those (or at least letting the user
known that they are missing something out).

Some examples:

- ensuring all container images are built from a particular base image
- guaranteeing that packages added have their `deb-src` counterpart providing
  the source code
- enforcing the use of non-root users
- avoiding leaving garbage from compile-time dependencies




```
    
    .-----------------------.
    |                       |
    |  concourse/concourse  |
    |                       | --> does it come from `ubuntu:bionic`?
    *-----------------------*     what's the version of each package?
                                  does it run as non-user?



```




## the bare minimum of building container images


If we erase from our minds the fact that `Dockerfile` is the de-facto way of
building container images today, we can summarize the creation of that image as:

1. bringing some content from a local directory
2. installing some dependencies
3. configuring some runtime parameters (environment variables, volumes, etc)



```

      local files -----------------.
                                   |
           +                       |
                                   |
      debian packages -------------+---(magic building)---==> container image
                                   |
           +                       |
                                   |
      container runtime configs ---*


```


i.e., regardless of how we get to that final container image, all we have to do
is get those pieces in, install some packages and configure some stuff.

Now, what it we tried to replicate what a Dockerfile does, but with `docker` and `bash`, instead?


## a primitive container image builder

Leveraging `docker` as the tool for running containers, we can achieve the same
as a Dockerfile:


```bash
# a *very* primitive "Dockerfile"-equivalent that
# is able to build a container image.
#
#
main () {
	local image

	image=$(docker pull ubuntu:bionic -q)
	image=$(add $image $(realpath ./file.txt) /file.txt)
	image=$(run $image 'apt-get update && apt-get install -y vim')
	image=$(entrypoint $image "/bin/echo")
	image=$(cmd $image "hello!!!")

	docker tag $image cirocosta/image
}
```

(see https://github.com/cirocosta/sample-manual-dockerfile)



```

FROM ubuntu:bionic


    1. pull initial root filesystem (ubuntu:bionic)



    FS_1: ubuntu

                    .--------------------
                    |
                    |    .
                    |    ├── etc
                    |    │   └── test.conf
                    |    └── var
                    |




ADD ./file.txt /file.txt


    1. create a container from a snapshot `FS_1`
      1.1 mount `file.txt` inside the container
    2. copy the file to the desired location


    FS_2 = snapshot(FS_1)


              
                    .--------------------
                    |
                    |    .
                    |    ├── etc
                    |    │   └── test.conf
                    |    ├── var
     file.txt-------+----└── mnt
                    |        └── file.txt (readonly mount)
                    |
                    |
                    |   --> /bin/sh -c cp /mnt/test.conf /file.txt
                    |
                    |



                    .--------------------
                    |
                    |    .
                    |    ├── etc
                    |    │   └── test.conf
                    |    ├── file.txt  <<<< new file! 
                    |    └── var
                    |


              >> we just mutated the filesystem.



RUN apt update && apt install -y vim



    FS_3 = snapshot(FS_2)


                    .--------------------
                    |
                    |    .
                    |    ├── etc
                    |    │   └── test.conf
                    |    ├── file.txt  <<<< new file! 
                    |    └── var
                    |
                    |
                    |   --> /bin/sh -c apt update && apt install -y vim
                    |


      
                ==> mutates the filesystem, adding the contents of the
                    vim package and some other `apt` stuff



                    .--------------------
                    |
                    |    .
                    |    ├── etc
                    |    │   └── test.conf
                    |    ├── file.txt
                    |    ├── usr                  //
                    |    │   └── bin              //  
                    |    │        └── vim         //   new!!!1
                    |    └── var
                    |        └── lib              //
                    |             └── apt         //
                    |                  └── lists  //  neww!!1
                    |



                ==> take a snapshot of this final stage and distribute it


```


Although this might seem quite primitive, the point is that all that it takes is
to have a way of:

1. creating a container from a given image
2. mutating that filesystem
3. taking a snapshot of that filesystem at a given point in time

Nest that multiple times, and one can have any Dockerfile built. 

The good about Dockerfiles though is that you don't need to know any of that
stuff - just like with a traditional compiler you don't need to care about the
bare details of crafting machine-readable code (you focus on the higher-level
language), Dockerfiles give you the same.



## compilers 

If we compare Dockerfiles to regular compiled programming languages, and for a
moment, assume that the process of going from a Dockerfile to a container image
is a compilation process, we could start drawing some quite neat analogies.

```

                   .------------.
                   |            |
      source ----> |  compiler  | ------>  output
                   |            |
                   *------------*

       c            C compiler           binary


```


As there's a clear distinction between the types of jobs that the compiler do in
different phases, and that innovation can happen differently in each front, some
compiler infrastructures like LLVM separate that quite clearly in two fronts,
with a common "lower level language" in the middle (the intermediary
representation - IR):

- the `frontend`, dealing with syntax and semantics
- the `backend`, dealing with the process of optimizing the code and creating
  the final output.



```

    source ---> frontend  ----> IR  ---> backend  ---> output


   C code       clang       LLVM IR      LLVM          binary

```


That way,  C language designers can develop code that makes the C language
specification move forwards, not having to deal with the dragons of optimizing
the compilation steps, alongside Fortran developers, who target the same
intermediary representation, that can benefit from the same backend.




```

    source ---> frontend  ----> IR  ---> backend  ---> output


   C code       clang       LLVM IR      LLVM          binary
   FORTRAN      flang
   Swift        ...
   Rust         ...
   ...          ...


```

An, that's what's hapenning right now for container images.



## buildkit

In 2017, some folks at Docker started working on ways of decoupling Docker's
build mechanism so that it could be iterated on in a faster way.


```
                            .-------> dockerd
  docker engine --> moby  --+-------> docker cli
                            *-------> buildkit
                            ... other components

```


Buildkit comes with the same mindset as LLVM - provide a common infrastructure
for building container image, separating the concerns of developing frontends
from the backend infrastructure.


PS.: and now we're at a turning point where we can focus on the innovation where
it matters, and don't need to reivent the wheel when it comes to creating those
container images


```


    source -------> frontend  ----> IR  ---> backend  ---> output



    Dockerfile    dockerfile.v0     LLB      buildkitd    containerimage

     

    |                           |    |                    |
    *---------------------------*    *--------------------*

     understanding what to build      figuring out how to build
                                       +  actually doing so

```

That backend infrastructure leverages those exact same three concepts that we talked
before when we were creating the primitive image builder - using containers that
run on top of snapshots to mutate their state and advance in the creation of a
final container image -, that is:

1. creating a container from a given image


```
  

  run (image + config) 


        ===> materialized as ==>


               container ----------------
               |  
               |  leveraging that filesystem
               |             +
               |     using that config


```


2. mutating that filesystem



```



  
    container ----                                     container ---------
    |                                                  |
    |   filesystem       ==> "run some commands ..."   |  modified filesystem
    |                                                  |
    

```




3. taking a snapshot of that filesystem at a given point in time


```


      container ---------
      |
      |  modified filesystem        ===> snapshot() ===>    image
      |
    


```



Aside from that, just like a compiler backend, it takes care of all of the
optimizations that would be hard for us to implement in our very primiter
builder - caching and running steps in parallel.


```


                        SYNTAX                 --------.
                           ===>                        |
                                                       |
                             FRONTEND                  |  what we care about
                                  ===>                 |
                                                       |
             .------                LLB        --------*
             |                        ===> 
 buildkitd   |
             |                          STATE MUTATIONS 
             *------                                 ===> final container image



```


## a minimally viable frontend for our container images

With that idea of what most of our container images are, I went on defining a
syntax that could represent those concepts.


```


      local files -----------------.
                                   |
           +                       |
                                   |
      debian packages -------------+---(magic building)---==> container image
                                   |
           +                       |
                                   |
      container runtime configs ---*


```

To build that syntax up right from actual needs, we can break those down into
certain "content installation steps":

- setting up a base image
- installing debian packages
- getting local files in
- setting container runtime configurations


## gathering the base image

As Concourse always leverages a base image that is not just plain scratch, it
made sense for me to start with the retrieval of an image that the user
specifies to move forward with the mutations later on.


```hcl
# declaration of how we want the final image to
# look like once it gets built.
#
image "cirocosta/sample" {

  # the base image that powers the final container
  # image that we're building.
  #
  base_image {
    name = "ubuntu"
    ref  = "bionic"
  }

}
```

Putting that through what would be the equivalent of a compilation pipeline,
we'd translate such syntax into LLB in our frontend:


```

      estaleiro.hcl  ===>  estaleiro frontend ==> LLB  ==> buildkitd ==> image


```

Looking at the LLB generated:


```yaml
op:
  source.image: "docker-image://docker.io/library/ubuntu:latest@sha256:c303f19cfe9ee92badbbbd7567bc1ca47789f79303ddcef56f77687d4744cd7a"
  digest: "sha256:9d0df2288a6f52c368c12a5fc188130f6db0125cd446bbf882cb2d513760bc73"

op:
  source.image: "docker-image://docker.io/cirocosta/estaleiro@sha256:f27dd2a0011116a05346f966c79699a0bb10ff197240af3d90efd11543dfa43a"
  digest: "sha256:d9e85dc882e618445ae8164dd5dee13d7ba8bc9b18487cca4338e5fa51aa3913"

op:
  inputs:
    - sha256:9d0df2288a6f52c368c12a5fc188130f6db0125cd446bbf882cb2d513760bc73
    - sha256:d9e85dc882e618445ae8164dd5dee13d7ba8bc9b18487cca4338e5fa51aa3913
  exec:
    args: ["/usr/local/bin/estaleiro", "collect", "-i=/var/lib/dpkg/status"]
  mounts:
    - input: 0
      dest: "/"
    - input: 1
      dest: "/usr/local/bin/estaleiro"
  digest: "sha256:155669101ce3ce82852d075af1a68fb32730100c720993b764fbfac886dffe12"

op:
  inputs: 
    - sha256:155669101ce3ce82852d075af1a68fb32730100c720993b764fbfac886dffe12
  digest: "sha256:031e04205cbb52e5ad87530ca0fc659586048bf8bc7b028d47397fd4a3cf6fc8"
```

We could rewrite that to some more simpler terms:

```


op:
  source.image: "docker-image://docker.io/library/ubuntu:latest"
  digest: "ubuntu"

op:
  source.image: "docker-image://docker.io/cirocosta/estaleiro"
  digest: "estaleiro"

op:
  inputs:
    - ubuntu
    - estaleiro
  exec:
    args: ["/usr/local/bin/estaleiro", "collect", "-i=/var/lib/dpkg/status"]
  mounts:
    - input: 0
      dest: "/"
    - input: 1
      dest: "/usr/local/bin/estaleiro"
  digest: "modified-fs"

op:
  inputs: 
    - modified-fs
  digest: "final-digest"
```

As that all boils down to a directed acyclic graph (DAG), we can represent that
as such:


![](./graph.svg)


## installing packages



## container image as an artifact

two concepts:

- sources
  - base image
  - static
    - files
    - tarballs
  - dynamic
    - packages
- runtime configuration



## what enabled this

## from syntax to container image

### snapshots


## concourse images as an example

Concourse has quite a few number of images to build:

- `concourse/concourse`
- `concourse/*-resource` (13 of them)

This means that there's a lot of source code to be scanned by the OSLO
team to ensure that we're not shipping licenses that hurt our customers.



## references


- [alternatives to distroless requirements and recommendations](https://docs.google.com/document/d/14jYh0YpCJ2NNYXzSye46JsL8eSFRgysP6TVeSZu7SD4/edit) 
