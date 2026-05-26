GO      ?= go
NAME    := mossspore
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-X github.com/moss/mossspore/internal/version.Version=$(VERSION) -X github.com/moss/mossspore/internal/version.Commit=$(COMMIT) -X github.com/moss/mossspore/internal/version.Date=$(DATE)"

BINDIR := build

.PHONY: all build build-linux build-windows build-darwin clean test lint run

all: build

build:
	$(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)$(shell go env GOEXE) ./cmd/$(NAME)

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-linux-amd64 ./cmd/$(NAME)

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-windows-amd64.exe ./cmd/$(NAME)

build-darwin:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-darwin-arm64 ./cmd/$(NAME)

clean:
	rm -rf $(BINDIR)

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

run: build
	$(BINDIR)/$(NAME)$(shell go env GOEXE) $(ARGS)
