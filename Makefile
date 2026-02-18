BINARY  := machinist
MODULE  := github.com/moinsen-dev/machinist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.Version=$(VERSION)

.PHONY: build clean install test

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/machinist

clean:
	rm -f $(BINARY)

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/machinist

test:
	go test ./... -race -count=1
