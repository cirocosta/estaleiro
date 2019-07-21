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
		--frontend=gateway.v0 \
		--local context=. \
		--local dockerfile=.


image-frontend:
	docker build --target frontend -t cirocosta/estaleiro-frontend .

run-buildkitd:
	docker run -d --privileged -p 1234:1234 moby/buildkit:latest --addr tcp://0.0.0.0:1234
