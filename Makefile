build:
	go build -i -v .

image-frontend:
	docker build --target frontend -t cirocosta/estaleiro-frontend .

image-estaleiro:
	DOCKER_BUILDKIT=1 docker build -f ./estaleiro.hcl .

test:
	go test -v ./...
