FROM golang:alpine

COPY . /go/src/github.com/ubergesundheit/assets-server

RUN apk --no-cache --virtual .build add git bash alpine-sdk cmake && \
  go get -u -v github.com/rakyll/statik \
    github.com/lpar/gzipped \
    github.com/otiai10/copy && \
  go get -u -d -v github.com/google/brotli/go/cbrotli && \
  cd $GOPATH/src/github.com/google/brotli && \
  mkdir out && \
  cd out && \
  ../configure-cmake --disable-debug && \
  cmake .. && \
  make -j8 && \
  make test && \
  make install && \
  go install github.com/ubergesundheit/assets-server/cmd/as-builder && \
  apk del .build
