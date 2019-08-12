GIT_COMMIT = $(shell git rev-parse HEAD)

all: install linux-binary


install:
	go install -tags "dfrunmount local" -v .


test:
	go test -v ./...


llb-build: install linux-binary
	estaleiro llb \
                  --filename ./estaleiro.hcl \
                  --var estaleiro-commit:$(git rev-parse HEAD) \
		  | buildctl build \
		  	--local context=. \
		  	--local bin=bin \
			--output type=image,name=docker.io/cirocosta/estaleiro-test,push=true


linux-binary:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 \
		go build -v \
			-o ./bin/estaleiro \
			-tags "dfrunmount" \
			-ldflags "-X main.version=$(shell cat ./VERSION) -extldflags \"-static\"" \
			.


ubuntu:
	DOCKER_BUILDKIT=1 \
		docker build \
			--tag cirocosta/estaleiro-with-ubuntu \
			--target release \
			.
	docker run \
		--interactive \
		--tty \
		--entrypoint /bin/bash \
		cirocosta/estaleiro-with-ubuntu


# TODO include estaleiro-commit through `--var`
# TODO make it daemonless (`img`-like)
#
run:
	estaleiro build \
		--filename estaleiro.hcl \
		--local context:. \
		--local dockerfile:.


llb:
	@estaleiro llb -f ./estaleiro.hcl --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl debug dump-llb \
		| jq '.'


graph: install
	estaleiro llb -f ./estaleiro.hcl --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl debug dump-llb --dot \
		| dot -Tsvg > graph.svg
	open -a "Firefox" ./graph.svg


buildctl-gateway-integration:
	buildctl build \
		--frontend gateway.v0 \
		--local context=. \
		--local dockerfile=. \
		--opt build-arg:estaleiro-commit=$(GIT_COMMIT) \
		--opt source=cirocosta/estaleiro \
		--trace /tmp/trace \
		--output type=oci,dest=image.tar


docker-integration:
	DOCKER_BUILDKIT=1 \
		docker build \
			--tag a \
			--build-arg estaleiro-commit=$(GIT_COMMIT) \
			--file ./estaleiro.hcl \
			.


image:
	DOCKER_BUILDKIT=1 \
		docker build \
			--target frontend \
			--tag cirocosta/estaleiro \
			.


run-buildkitd:
	docker run \
		--detach \
		--privileged \
		--publish 1234:1234 \
		moby/buildkit:latest \
		--addr tcp://0.0.0.0:1234
