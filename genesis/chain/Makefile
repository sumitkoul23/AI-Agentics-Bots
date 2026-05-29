BINARY = agenticd
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.1-genesis")
LD_FLAGS = -X github.com/cosmos/cosmos-sdk/version.Name=agentic \
           -X github.com/cosmos/cosmos-sdk/version.AppName=$(BINARY) \
           -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION)

.PHONY: all install build test lint proto clean

all: build

install:
	go install -ldflags "$(LD_FLAGS)" ./cmd/$(BINARY)

build:
	go build -ldflags "$(LD_FLAGS)" -o ./build/$(BINARY) ./cmd/$(BINARY)

test:
	go test -race -timeout 5m ./...

lint:
	golangci-lint run --timeout 5m ./...

proto:
	./scripts/protogen.sh

clean:
	rm -rf ./build
