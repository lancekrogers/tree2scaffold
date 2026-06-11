# Build recipes — local and cross-platform variants
binary_name := "tree2scaffold"
main_path := "./cmd/tree2scaffold"
bin_dir := "bin"
ldflags := "-s -w"

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
