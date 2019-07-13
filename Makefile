build:
	go build -i -v .

image:
	docker build -t cirocosta/estaleiro .

test:
	go test -v ./...
