# Test recipes

# Show available test commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run unit + integration tests
[no-cd]
all: unit integration
    @echo "✅ All tests passed"

# Run unit tests with coverage
[no-cd]
unit:
    go test -cover ./...

# Run unit tests with verbose output
[no-cd]
unit-verbose:
    go test -v -cover ./...

# Run the end-to-end integration tests
[no-cd]
integration:
    TEST_ALL=1 go test -timeout 30s -v ./test

# Generate HTML coverage report
[no-cd]
coverage-html:
    go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Run the WASI integration test (skips if no runtime; T2S_WASM reuses a prebuilt module)
[no-cd]
wasm:
    T2S_WASM="$PWD/bin/tree2scaffold.wasm" TEST_ALL=1 go test -timeout 120s -v -run WASM ./test
