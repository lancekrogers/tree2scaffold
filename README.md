# tree2scaffold (t2s)

An open‑source CLI tool that converts an ASCII `tree`‑style project layout into a real directory and file scaffold, with support for language‑specific file generators. Perfect for rapidly implementing project structures from LLM conversations.

---

## Features

- **ASCII‑to‑Scaffold**: Paste or pipe in a `tree` output and quickly generate the entire project structure.
- **Multiple Input Formats**:
  - Standard `tree` command output with ascii characters (├── and └──)
  - Directory structure with indentation and trailing slashes
  - Simple file list (one path per line)
- **Clipboard Fallback**: If you invoke `tree2scaffold` with no piped input, it automatically reads from the macOS clipboard (`pbpaste`).
- **Modular Generators**: File content is generated per‑extension:
  - **`.go`** files get a full stub with appropriate package name and structure:
    - `main.go` files always get `package main` and a `func main()` scaffold.
    - Other Go files get proper package name based on their directory.
  - All other extensions (e.g. `.py`, `.js`, `.md`, `.yaml`) get only a comment header, using the correct syntax for the filetype.
  - Easily extend via `RegisterGenerator(ext, genFunc)` in the content generator interface.
- **Intelligent File Handling**: Never overwrites existing files; only adds missing ones.
- **Structure Verification**: Validates that the generated structure matches the spec post-creation.
- **Preview & Confirm**: Use `-d` or `-dry-run` to see exactly which dirs/files will be created.
- **Progress Output**: Visual feedback for every `mkdir` and file write with colored symbols.
- **Cross‑Platform Design**: Written in Go, no external deps beyond standard Go and (optionally) `pbpaste` on macOS.

---

## Installation

### Quick Install (recommended)

```bash
# Build and copy the binary into /usr/local/bin
# This also creates the t2s alias automatically
make install
```

You can override the install location:

```bash
make install PREFIX="$HOME/.local"
```

### Via Go Modules

```bash
# This installs both tree2scaffold and the t2s alias
go install github.com/lancekrogers/tree2scaffold/cmd/tree2scaffold@latest
```

Ensure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH`.

---

## Usage

You can use either the `tree2scaffold` command or the shorter `t2s` alias for any command:

```bash
# Create a project from a tree specification in your clipboard
t2s -root ./myproject

# Or pipe a tree spec directly
pbpaste | t2s -root ./myproject

# Or just run in an empty dir (clipboard fallback):
cd ~/Projects/NewApp
t2s

# Preview only, then confirm (shortcut flag):
t2s -d

# Skip confirmation prompt:
t2s -yes
```

### Command-line Flags

- `-root <path>`: Directory under which to build the scaffold (defaults to `.`).
- `-d`, `-dry-run`: Show what would be created and prompt for confirmation, without writing.
- `-yes`: Skip the confirmation prompt (useful for scripts).
- `-force`: Force overwrite of files that conflict with directories.
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

The tool uses an interface-based approach for content generation that makes it easy to extend:

```go
// ContentGenerator generates content for files
type ContentGenerator interface {
    // GenerateContent creates content for a file based on its path and comment
    GenerateContent(relPath string, comment string) string
    
    // RegisterGenerator adds a new generator for a specific extension or filename
    RegisterGenerator(extOrName string, generator FileGenerator)
}
```

### Method 1: Registering with the Default Generator

The simplest way to add custom generators is to use the default implementation:

```go
// Create a new default content generator
generator := scaffold.NewDefaultContentGenerator()

// Register a custom generator for Python files
generator.RegisterGenerator(".py", func(path, comment string) string {
    header := ""
    if comment != "" {
        header = fmt.Sprintf("# %s\n", comment)
    }
    return header + `def main():\n    pass\n\nif __name__ == "__main__":\n    main()\n`
})

// Create a scaffolder that uses your generator
scaffolder := &scaffold.DefaultScaffolder{
    ForceMode:       false,
    ContentProvider: generator,
}
```

### Method 2: Implementing Your Own Content Generator

For more complex customization, you can implement the ContentGenerator interface:

```go
type MyContentGenerator struct {
    // Your custom state here
}

func (g *MyContentGenerator) GenerateContent(relPath, comment string) string {
    // Your custom content generation logic
    // ...
}

func (g *MyContentGenerator) RegisterGenerator(extOrName string, generator scaffold.FileGenerator) {
    // Register a custom generator
    // ...
}
```

---

## Testing

The project includes comprehensive testing to ensure reliability:

- **Unit tests**: Test individual components

  ```bash
  go test ./...
  ```

- **Integration tests**: Test end-to-end behavior with multiple scenarios

  ```bash
  make integration
  ```
  
  This runs comprehensive tests that verify the complete functionality by:
  
  - Testing simple project structures
  - Testing complex nested structures with hundreds of files
  - Verifying hidden directory handling (.github, .vscode)
  - Testing cross-platform file recognition (Windows, Linux, macOS)
  - Using checksums to verify structure integrity
  
- **Run all tests**:

  ```bash
  make test-all
  ```

---

## Using with AI Models

`tree2scaffold` (or `t2s`) is specifically designed to work with AI-generated project structures from tools like ChatGPT and Claude. Here's how to use it effectively:

### With ChatGPT / Claude

1. Ask the AI to generate a project structure in ASCII tree format:
   ```
   Generate a tree structure for a simple Flask web application with templates, static files, and a database
   ```

2. Copy the ASCII tree output from the AI response

3. Run `t2s` in your desired project directory:
   ```bash
   cd ~/Projects/NewFlaskApp
   t2s -d  # Preview first
   ```

### With Neovim Integration

For Neovim users, you can create a simple command to pipe the current buffer or selection to `t2s`:

```lua
vim.api.nvim_create_user_command('Tree2Scaffold', function(opts)
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  local text = table.concat(lines, '\n')
  local tmp = os.tmpname()
  local f = io.open(tmp, 'w')
  f:write(text)
  f:close()
  
  local cmd = string.format('cat %s | t2s -root .', tmp)
  vim.fn.system(cmd)
  os.remove(tmp)
  
  print("Scaffolding complete!")
end, {})
```

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
