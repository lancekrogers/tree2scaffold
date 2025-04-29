# Makefile for tree2scaffold

# automatically pick up module name from go.mod
MODULE := $(shell go list -m)
BINARY := tree2scaffold
CMD := ./cmd/tree2scaffold
GO := go

.PHONY: all build install test fmt lint clean help

# Default: run tests, then build
all: test build

# Build the CLI binary into ./bin/
build:
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY) $(CMD)

# Install into your $GOPATH/bin or $GOBIN
install:
	$(GO) install $(CMD)

# Run all tests
test:
	$(GO) test ./...

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
	@echo "  make          → runs tests, then builds"
	@echo "  make build    → compile binary to ./bin/$(BINARY)"
	@echo "  make install  → go install $(MODULE)/cmd/$(BINARY)"
	@echo "  make test     → run all tests"
	@echo "  make fmt      → run go fmt ./..."
	@echo "  make lint     → run golangci-lint"
	@echo "  make clean    → remove ./bin"
