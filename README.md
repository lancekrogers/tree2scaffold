# tree2scaffold

An open‑source CLI tool that converts an ASCII `tree`‑style project layout into a real directory and file scaffold, with support for language‑specific file generators.

---

## Features

- **ASCII‑to‑Scaffold**: Paste or pipe in a `tree` output and quickly generate the entire project structure.
- **Multiple Input Formats**:
  - Standard `tree` command output with ascii characters (├── and └──)
  - Directory structure with indentation and trailing slashes
  - Simple file list (one path per line)
- **Clipboard Fallback**: If you invoke `tree2scaffold` with no piped input, it automatically reads from the macOS clipboard (`pbpaste`).
- **Modular Generators**: File content is generated per‑extension:
  - **`.go`** files get a full stub with `package <name>` and a `func main()` scaffold.
  - All other extensions (e.g. `.py`, `.js`, `.md`, `.yaml`) get only a comment header, using the correct syntax for the filetype.
  - Easily register new generators in the `init()` function in `pkg/scaffold/generators.go` via `RegisterGenerator(ext, genFunc)`.
- **Preview & Confirm**: Use `-dry-run` to see exactly which dirs/files will be created, then confirm before making changes (or skip with `-yes`).
- **Progress Output**: Visual feedback for every `mkdir` and file write with colored symbols.
- **Cross‑Platform Design**: Written in Go, no external deps beyond standard Go and (optionally) `pbpaste` on macOS.

---

## Installation

### Via Makefile (recommended)

```bash
# Build and copy the binary into /usr/local/bin (requires sudo if needed)
make install
```

You can override the install location:

```bash
make install PREFIX="$HOME/.local"
```

### Via Go modules

```bash
go install github.com/lancekrogers/tree2scaffold/cmd/tree2scaffold@latest
```

Ensure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH`.

---

## Usage

```bash
# Pipe a tree spec into scaffold (explicit root)
pbpaste | tree2scaffold -root ./myproject

# Or just run in an empty dir (clipboard fallback):
cd ~/Projects/NewApp
tree2scaffold

# Preview only, then confirm:
tree2scaffold -dry-run

# Skip confirmation prompt:
tree2scaffold -yes
```

### Flags

- `-root <path>`: Directory under which to build the scaffold (defaults to `.`).
- `-dry-run`: Show what would be created and prompt for confirmation, without writing.
- `-yes`: Skip the confirmation prompt (useful for scripts).
- `-debug`: Output additional debug information.

### Input Format Examples

You can use any of these formats:

1. **Standard tree command output**:
```
myproject/
├── cmd/
│   └── app.go
└── pkg/
    └── utils.go
```

2. **Simple file list**:
```
cmd/
cmd/app.go
pkg/
pkg/utils.go
```

3. **Alternative tree format** (with or without trailing slashes for directories):
```
myproject/
├── cmd/
├── internal/
│   ├── api/
│   └── db/
└── pkg/
```

---

## Customizing File Generators

By default, generators are registered in `pkg/scaffold/generators.go`:

```go
func init() {
    RegisterGenerator(".go", generateGo)         // Go stub generator
    // other extensions fall back to default comment generator
}
```

To add support for, say, Python or Rust:

```go
RegisterGenerator(".py", func(path, comment string) string {
    header := ""
    if comment != "" {
        header = fmt.Sprintf("# %s\n", comment)
    }
    return header + `def main():\n    pass\n`
})
```

---

## Testing

- **Unit tests:**

  ```bash
  go test ./...
  ```

- **Integration test:**
  ```bash
  make integration
  ```
  
  This runs a comprehensive test that verifies the complete functionality by creating a scaffold from an example tree structure.

---

## Contributing

Contributions are welcome! Please open issues or pull requests on [GitHub](https://github.com/lancekrogers/tree2scaffold).

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/...`)
3. Write tests and update code
4. Submit a PR

Please follow standard Go formatting and linting:

```bash
go fmt ./...
golangci-lint run
```

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
