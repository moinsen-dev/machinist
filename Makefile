BINARY  := machinist
MODULE  := github.com/moinsen-dev/machinist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.Version=$(VERSION)

.PHONY: build clean install test release-local release

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/machinist

clean:
	rm -f $(BINARY)
	rm -rf dist/

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/machinist

test:
	go test ./... -race -count=1

# Build release archives locally (no publish) — useful for testing
release-local:
	goreleaser release --snapshot --clean

# Tag and release via goreleaser (pushes to GitHub)
release:
	@if [ -z "$(TAG)" ]; then echo "Usage: make release TAG=v0.1.0"; exit 1; fi
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
	goreleaser release --clean
