# Build recipes — local and cross-platform variants
binary_name := "tree2scaffold"
main_path := "./cmd/tree2scaffold"
bin_dir := "bin"
ldflags := "-s -w"
wasm_runtime := env_var_or_default("WASMTIME", "wasmtime")

# Show available build commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Build the binary into bin/ (default)
[no-cd]
local:
    @mkdir -p {{bin_dir}}
    go build -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Quick development build (no optimization flags)
[no-cd]
dev:
    @mkdir -p {{bin_dir}}
    go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

# --- Cross-platform builds ---

# Build for Linux AMD64
[no-cd]
linux:
    @mkdir -p {{bin_dir}}
    GOOS=linux GOARCH=amd64 go build -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

# Build for macOS Intel
[no-cd]
darwin:
    @mkdir -p {{bin_dir}}
    GOOS=darwin GOARCH=amd64 go build -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

# Build for macOS Apple Silicon
[no-cd]
darwin-arm64:
    @mkdir -p {{bin_dir}}
    GOOS=darwin GOARCH=arm64 go build -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

# Build for Windows AMD64
[no-cd]
windows:
    @mkdir -p {{bin_dir}}
    GOOS=windows GOARCH=amd64 go build -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}}-windows-amd64.exe {{main_path}}

# Build for all platforms
all: linux darwin darwin-arm64 windows
    @echo "All platform builds complete"

# --- WebAssembly (WASI / wasip1) ---
# Opt-in: intentionally NOT part of `all`, so it only runs when requested.

# Build the WASI artifact (standard Go, GOOS=wasip1)
[no-cd]
wasm:
    @mkdir -p {{bin_dir}}
    GOOS=wasip1 GOARCH=wasm go build -trimpath -ldflags '{{ldflags}}' -o {{bin_dir}}/{{binary_name}}.wasm {{main_path}}

# Portability gate — fail fast on non-wasip1-safe code
[no-cd]
wasm-vet:
    GOOS=wasip1 GOARCH=wasm go vet ./...

# Smoke-run the wasm against the current dir (pipe a tree on stdin)
[no-cd]
wasm-run: wasm
    {{wasm_runtime}} run --dir .::/ --env PWD=/ {{bin_dir}}/{{binary_name}}.wasm -- -root . -yes

# Release artifact + SHA-256 checksum
[no-cd]
wasm-release: wasm
    cd {{bin_dir}} && shasum -a 256 {{binary_name}}.wasm > {{binary_name}}.wasm.sha256
