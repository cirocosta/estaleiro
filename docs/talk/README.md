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


With a single goal in mind - shipping Concourse container images to customers
running Kubernetes -, we can see that there's a problem to solve right in the
bottom of the preparation of our artifact to be distributed:




1. to get that to customers, it needs to be on PivNet,
  - so that we can attest what has been distributed to who

2. to have it on PivNet, it needs an OSL file
  - so that we can disitribute the copyright and prove that we don't have any
    licenses that would hurt our customers

3. to get the OSL file, it needs to let norsk know where to scan source code
  - so that it can know what are those licenses, and gather copyright info

4. to let Norsk know where to scan, it needs to provide a bill of materials


Naturally, if the bottommost - generating a container image that is scannable -
impacts the rest of the process, that means that we can't get the great features
that our teams produce to the hand of our customers.


To get real, let's look at one of the container images that we need to scan to
ship Concourse.



## the Concourse container image

As an example of a container image that we'd like to ship, let's look at how
`concourse/concourse` is created (with few modifications):


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


If we erase from our minds the fact that `Dockerfile` is the de-facto way of
building container images today, we can summarize the creation of that image as:

1. bringing some content from a local directory
2. installing some dependencies
3. configuring some runtime parameters (environment variables, volumes, etc)

i.e., regardless of how we get to that final container image, all we have to do
is get those pieces in, install some packages and configure some stuff.

What it we tried to replicate what a Dockerfile does with bash?


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

Although that's quite primitive, the point is that all that it takes is to have
a way of:

1. creating a container from a given image
2. mutating that filesystem
3. taking a snapshot of that filesystem at a given point in time

Nest that multiple times, and one can have any Dockerfile built. 

The good about Dockerfiles though is that you don't need to know any of that
stuff - just use the Dockerfile syntax and you're good.




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


