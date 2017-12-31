# assets-server
[![Docker Repository on Quay](https://quay.io/repository/geraldpape/as-builder/status "Docker Repository on Quay")](https://quay.io/repository/geraldpape/as-builder)

A tool to bundle a directory of files into a static HTTP server. Useful for serving html and js from Kubernetes Pods.

Uses https://github.com/rakyll/statik

## Use in a docker build stage

```Dockerfile
## First stage: Build production assets
FROM node:8-alpine as build

WORKDIR /usr/src/app
COPY package.json /usr/src/app/
RUN npm install
COPY . /usr/src/app
RUN npm run build

## Second stage: Build assets-server binary
FROM quay.io/geraldpape/as-builder:latest as packer

COPY --from=build /usr/src/app/build /assets

RUN as-builder -debug -src /assets -dest /assets-server -port 8080 -url /

## Final stage: Use static binary as small docker image
FROM scratch

COPY --from=packer /assets-server /assets-server

CMD ["/assets-server"]

```

## Installation

`go get github.com/ubergesundheit/assets-server/cmd/as-builder`

## Usage:

```
Usage of as-builder:
  -debug
    enable verbose debug messages
  -dest string
    file path of the resulting binary (default "assets-server")
  -port int
    TCP port from which the server will be reachable (default 8000)
  -src string
    file path of the assets directory (default "./public")
  -url string
    URL path for the server (default "/")
```
