GIT_COMMIT = $(shell git rev-parse HEAD)


build:
	go build -i -v .


test:
	go test -v ./...


run:
	./estaleiro build -f ./estaleiro.hcl --bom ./bom.yml --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl build --local context=.


graph:
	./estaleiro build -f ./estaleiro.hcl --bom ./bom.yml --var estaleiro-commit:$(GIT_COMMIT) \
		| buildctl debug dump-llb --dot \
		| dot -Tsvg > graph.svg


buildctl-integration:
	buildctl build \
		--frontend gateway.v0 \
		--opt source=cirocosta/estaleiro-frontend:rc \
		--local dockerfile=.


docker-integration:
	docker build \
		--tag a \
		--file ./estaleiro.hcl \
		.

image-frontend:
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
