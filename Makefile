GOOS ?= linux
GOARCH ?= amd64

ifeq ($(OS), Windows_NT)
	CURRENTDIR := $(shell cmd /c cd)
else
	CURRENTDIR := $(shell pwd)
endif

BUILD_IMAGE := pupok-proxy:latest

all: image build

image:
	docker build -t $(BUILD_IMAGE) .

build:
	docker run --rm -i \
		-v $(CURRENTDIR):/build/src/github.com/Andykaban/pupok-test-proxy \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-w /build/src/github.com/Andykaban/pupok-test-proxy \
		$(BUILD_IMAGE) go build
