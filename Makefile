build:
	go build -i -v .

image-frontend:
	docker build --target frontend -t cirocosta/estaleiro-frontend .

image-estaleiro:
	./estaleiro build -f ./estaleiro.hcl | buildctl build --local context=.

test:
	go test -v ./...

run-buildkitd:
	docker run -d --privileged -p 1234:1234 moby/buildkit:latest --addr tcp://0.0.0.0:1234
