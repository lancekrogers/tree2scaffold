# Makefile for tree2scaffold

# automatically pick up module name from go.mod
MODULE := $(shell go list -m)
BINARY := tree2scaffold
CMD := ./cmd/tree2scaffold
GO := go

.PHONY: all build install test integration fmt lint clean help

# Default: run unit tests, integration test, then build
all: test integration build

# Where to install your binary
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin

.PHONY: build install

build:
	@mkdir -p bin
	go build -o bin/tree2scaffold ./cmd/tree2scaffold

install: build
	@echo "Installing tree2scaffold to $(BINDIR)"
	@mkdir -p $(BINDIR)
	@cp bin/tree2scaffold $(BINDIR)/tree2scaffold
	@echo "Creating t2s alias -> tree2scaffold"
	@ln -sf $(BINDIR)/tree2scaffold $(BINDIR)/t2s


# Alternative: install directly via `go install`
install-go:
	$(GO) install $(MODULE)/cmd/tree2scaffold@latest
	@echo "Creating t2s alias -> tree2scaffold"
	@ln -sf $(shell go env GOPATH)/bin/tree2scaffold $(shell go env GOPATH)/bin/t2s

# Run all unit tests
test:
	$(GO) test -cover ./...

# Run the integration test (end-to-end CLI behavior)
integration:
	TEST_ALL=1 $(GO) test -timeout 30s -v ./test

# Format code (uses go fmt; change to goimports if you prefer)
fmt:
	$(GO) fmt ./...

# Lint (requires golangci-lint installed)
lint:
	golangci-lint run

# Remove built artifacts
clean:
	rm -rf bin

# Show available targets
help:
	@echo "Usage:"
	@echo "  make          → run tests, integration, then build"
	@echo "  make build    → compile binary to ./bin/$(BINARY)"
	@echo "  make install  → go install $(MODULE)/cmd/$(BINARY)"
	@echo "  make test     → run all unit tests with coverage"
	@echo "  make integration → run the end-to-end integration test"
	@echo "  make fmt      → run go fmt ./..."
	@echo "  make lint     → run golangci-lint"
	@echo "  make clean    → remove ./bin"
