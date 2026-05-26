GO      ?= go
NAME    := mossspore
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-X github.com/moss/mossspore/internal/version.Version=$(VERSION) -X github.com/moss/mossspore/internal/version.Commit=$(COMMIT) -X github.com/moss/mossspore/internal/version.Date=$(DATE)"

BINDIR := build

.PHONY: all build build-linux build-windows build-darwin build-all clean test lint run

all: build

build:
	$(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)$(shell go env GOEXE) ./cmd/$(NAME)

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-linux-amd64 ./cmd/$(NAME)

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-linux-arm64 ./cmd/$(NAME)

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-darwin-amd64 ./cmd/$(NAME)

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-darwin-arm64 ./cmd/$(NAME)

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINDIR)/$(NAME)-windows-amd64.exe ./cmd/$(NAME)

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

clean:
	rm -rf $(BINDIR)

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

run: build
	$(BINDIR)/$(NAME)$(shell go env GOEXE) $(ARGS)
