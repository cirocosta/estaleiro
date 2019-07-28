# syntax = docker/dockerfile:experimental

FROM golang AS golang


FROM golang AS base

	ENV CGO_ENABLED=0
	RUN apt update && apt install -y git

	ADD . /src
	WORKDIR /src


FROM base AS build

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


FROM ubuntu AS ubuntu-with-estaleiro

	RUN set -x && \
		sed -i 's/# deb-src/deb-src/g' /etc/apt/sources.list && \
		apt update

	COPY \
		--from=build \
		/bin/estaleiro \
		/usr/local/bin/estaleiro
