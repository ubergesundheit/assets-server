# assets-server

A tool to bundle a directory of files into a static HTTP server. Useful for serving html and js from Kubernetes Pods.

Uses https://github.com/rakyll/statik

## Installation

`go get github.com/ubergesundheit/assets-server/cmd/as-builder`

## Usage:

```
Usage of as-builder:
  -dest string
    file path of the resulting binary (default "assets-server")
  -port int
    TCP port from which the server will be reachable (default 8000)
  -src string
    file path of the assets directory (default "./public")
  -url string
    URL path for the server (default "/")
```
