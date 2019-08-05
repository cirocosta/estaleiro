GIT_COMMIT = $(shell git rev-parse HEAD)


install:
	go install -tags "dfrunmount" -v .



image: install linux-binary
	estaleiro llb \
                  --filename ./estaleiro.hcl \
                  --var estaleiro-commit:$(git rev-parse HEAD) \
		  | buildctl build \
		  	--local context=. \
		  	--local bin=bin \
			--output type=image,name=docker.io/cirocosta/estaleiro,push=true


linux-binary:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 \
		go build \
			-o ./bin/estaleiro \
			-tags "dfrunmount" \
			-ldflags "-X main.version=$(shell cat ./VERSION) -extldflags \"-static\"" \
			.


test:
	go test -v ./...


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
run:
	estaleiro build \
		--filename estaleiro.hcl \
		--local context:. \
		--local dockerfile:.


llb:
	@estaleiro llb -f ./estaleiro.hcl --bom ./bom.yml --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl debug dump-llb \
		| jq '.'


graph: install
	estaleiro llb -f ./estaleiro.hcl --bom ./bom.yml --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl debug dump-llb --dot \
		| dot -Tsvg > graph.svg
	open -a "Firefox" ./graph.svg


buildctl-gateway-integration:
	buildctl build \
		--frontend gateway.v0 \
		--opt source=cirocosta/estaleiro-frontend:rc \
		--local dockerfile=.


docker-integration:
	DOCKER_BUILDKIT=1 \
		docker build \
			--tag a \
			--build-arg estaleiro-commit=$(GIT_COMMIT) \
			--file ./estaleiro.hcl \
			.


image-frontend:
	DOCKER_BUILDKIT=1 \
		docker build \
			--target frontend \
			--tag cirocosta/estaleiro-frontend:rc \
			.


run-buildkitd:
	docker run \
		--detach \
		--privileged \
		--publish 1234:1234 \
		moby/buildkit:latest \
		--addr tcp://0.0.0.0:1234
