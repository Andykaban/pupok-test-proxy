GOPATH=$(shell pwd)/vendor:$(shell pwd)

build:
	@GOPATH=$(GOPATH) go build
