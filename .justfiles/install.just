# Install recipes — install tree2scaffold and the t2s alias
binary_name := "tree2scaffold"
alias_name := "t2s"
bin_dir := "bin"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")

# Show available install commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Build and install to $GOBIN (default)
[no-cd]
local:
    #!/usr/bin/env bash
    set -euo pipefail
    just build local
    just install _copy

# Install from current bin/ (whatever was last built)
[no-cd]
current:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -f {{bin_dir}}/{{binary_name}} ]; then
        echo "Error: no binary found at {{bin_dir}}/{{binary_name}}"
        echo "Run 'just build local' first"
        exit 1
    fi
    just install _copy

# Install directly via `go install` (+ t2s alias)
[no-cd]
go:
    #!/usr/bin/env bash
    set -euo pipefail
    go install $(go list -m)/cmd/{{binary_name}}@latest
    ln -sf {{gobin}}/{{binary_name}} {{gobin}}/{{alias_name}}
    echo "{{binary_name}} installed to {{gobin}} (alias: {{alias_name}})"

# Copy binary to $GOBIN, create t2s alias, and sign on macOS
[no-cd, private]
_copy:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing {{binary_name}}..."
    mkdir -p {{gobin}}
    cp {{bin_dir}}/{{binary_name}} {{gobin}}/{{binary_name}}
    ln -sf {{gobin}}/{{binary_name}} {{gobin}}/{{alias_name}}
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "Signing {{binary_name}} binary for macOS..."
        codesign --force --sign - {{gobin}}/{{binary_name}} 2>/dev/null || \
        echo "Warning: Could not sign binary (non-fatal)"
    fi
    echo "{{binary_name}} installed to {{gobin}}/{{binary_name}} (alias: {{alias_name}})"
