# syntax = docker/dockerfile:experimental

FROM golang AS golang


FROM golang AS base

	ENV CGO_ENABLED=0
	RUN apt update && apt install -y git



FROM base AS build
	ADD ./go.mod /src/go.mod
	ADD ./go.sum /src/go.sum
	WORKDIR /src

	RUN go mod download

	ADD . /src

	RUN \
		--mount=type=cache,target=/root/.cache/go-build \
		--mount=type=cache,target=/go/pkg/mod \
			go build -v \
			-tags "netgo dfrunmount" \
			-o /bin/estaleiro \
			-ldflags "-X main.version=$(cat ./VERSION) -extldflags \"-static\""


FROM base AS test

	RUN go test -v ./...


FROM ubuntu AS release

	COPY \
		--from=build \
		/bin/estaleiro \
		/usr/local/bin/estaleiro

	ENTRYPOINT [ "/usr/local/bin/estaleiro" ]


FROM release AS frontend

	ENTRYPOINT [ "/usr/local/bin/estaleiro", "frontend" ]
