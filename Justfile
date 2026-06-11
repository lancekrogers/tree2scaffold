#!/usr/bin/env just --justfile
# tree2scaffold build and development tasks

set dotenv-load := true

# Configuration
binary_name := "tree2scaffold"
alias_name := "t2s"
bin_dir := "bin"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")

# Modules
[doc('Build the binary (local + cross-platform)')]
mod build '.justfiles/build.just'

[doc('Testing (unit, integration, coverage)')]
mod test '.justfiles/test.just'

[doc('Install tree2scaffold (+ t2s alias) to $GOBIN')]
mod install '.justfiles/install.just'

[doc('Linting (golangci-lint, vet)')]
mod lint '.justfiles/lint.just'


[private]
default:
    #!/usr/bin/env bash
    echo "tree2scaffold - ASCII tree to filesystem scaffolder"
    echo ""
    just --list --unsorted

# Format Go code
fmt:
    go fmt ./...

# Remove build artifacts
clean:
    rm -rf {{bin_dir}}

# Update and tidy dependencies
deps:
    go get -u ./...
    go mod tidy

# Uninstall tree2scaffold (+ t2s alias) from $GOBIN
uninstall:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Uninstalling tree2scaffold..."
    for f in {{binary_name}} {{alias_name}}; do
        if [ -e "{{gobin}}/$f" ]; then
            rm "{{gobin}}/$f"
            echo "removed {{gobin}}/$f"
        fi
    done
