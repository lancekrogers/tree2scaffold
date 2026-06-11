# Lint recipes

# Show available lint commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run golangci-lint on all packages
[no-cd]
all:
    golangci-lint run ./...

# Run go vet on all packages
[no-cd]
vet:
    go vet ./...

# Install required linting tools
[no-cd]
install-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
