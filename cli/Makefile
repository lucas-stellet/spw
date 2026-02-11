MODULE := github.com/lucas-stellet/spw
VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

.PHONY: build test lint clean snapshot install

build:
	go build -ldflags '$(LDFLAGS)' -o bin/spw ./cmd/spw/

test:
	go test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

snapshot:
	goreleaser release --snapshot --clean

install: build
	cp bin/spw $(GOPATH)/bin/spw 2>/dev/null || cp bin/spw /usr/local/bin/spw
