build:
	go build -i -v .

run:
	./estaleiro build -f ./estaleiro.hcl --bom ./bom.yml \
		| buildctl build --local context=.

graph:
	./estaleiro build -f ./estaleiro.hcl --bom ./bom.yml \
		| buildctl debug dump-llb --dot \
		| dot -Tsvg > graph.svg

image-frontend:
	docker build --target frontend -t cirocosta/estaleiro-frontend .

test:
	go test -v ./...

run-buildkitd:
	docker run -d --privileged -p 1234:1234 moby/buildkit:latest --addr tcp://0.0.0.0:1234
