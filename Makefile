GOPATH=$(shell go env GOPATH)
APP_NAME=go-resource.app
GIT_COMMIT=$(shell git rev-parse --short HEAD)

.PHONY: all test clean

build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $(APP_NAME) TheMain.go