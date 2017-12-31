FROM golang:alpine as build

COPY . /go/src/github.com/ubergesundheit/assets-server

RUN  apk --no-cache --virtual .build add git && \
  go get -u github.com/rakyll/statik && \
  go install github.com/ubergesundheit/assets-server/cmd/as-builder && \
  apk del .build
