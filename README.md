# assets-server
[![Docker Repository on Quay](https://quay.io/repository/geraldpape/as-builder/status "Docker Repository on Quay")](https://quay.io/repository/geraldpape/as-builder)

A tool to bundle a directory of files into a static HTTP server. Useful for serving html and js from Kubernetes Pods.

Automatically adds .gz and .br compressed files and serves it to clients supporting gz or brotli thanks to a modified version of https://github.com/lpar/gzipped.

Uses
- github.com/rakyll/statik
- github.com/lpar/gzipped
- github.com/google/brotli/go/cbrotli
- github.com/otiai10/copy
- github.com/amalfra/etag
- github.com/golang/gddo/httputil/header

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

## Manual Installation

Although possible, using the container image is recommended.
You'll need some build dependencies for brotli. (On Ubuntu: build-essential, cmake)

    go get -u github.com/rakyll/statik github.com/golang/gddo/httputil/header github.com/otiai10/copy github.com/amalfra/etag
    go get -u -d github.com/google/brotli/go/cbrotli
    cd $GOPATH/src/github.com/google/brotli
    mkdir out
    cd out
    ../configure-cmake --disable-debug
    cmake ..
    make -j8
    make test
    make install
    go get -u github.com/ubergesundheit/assets-server/cmd/as-builder

## Usage:

```
Usage of as-builder:
  -compress string
    comma separated list of file extensions to compress. To completely disable compression specify an empty string (default ".html,.htm,.css,.js,.svg,.json,.txt,.xml,.yml,.yaml,.kml,.csv,.tsv,.webmanifest,.vtt,.vcard,.vcf,.ttc,.ttf,.rdf,.otf,.appcache,.md,.mdown,.m3u,.m3u8")
  -debug
    enable verbose debug messages
  -dest string
    file path of the resulting binary (default "assets-server")
  -logging
    enable request logging for the server
  -port int
    TCP port from which the server will be reachable (default 8000)
  -src string
    file path of the assets directory (default "./public")
  -url string
    URL path for the server (default "/")
```
