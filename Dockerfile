FROM golang:alpine

COPY . /go/src/github.com/ubergesundheit/assets-server

ADD https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz /usr/local

RUN apk --no-cache --virtual .build add git bash alpine-sdk cmake xz && \
  go get -u -v github.com/rakyll/statik \
    github.com/golang/gddo/httputil/header \
    github.com/otiai10/copy \
    github.com/amalfra/etag && \
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
  xz -d -c /usr/local/upx-3.95-amd64_linux.tar.xz | \
  tar -xOf - upx-3.95-amd64_linux/upx > /bin/upx && \
  chmod a+x /bin/upx && \
  apk del .build && \
  apk --no-cache add binutils
