# tree2scaffold (`t2s`)

**The LLM gave you a tree. This makes the files.**

Paste an ASCII `tree`, a file list, or whatever ChatGPT boxed in a code
fence. `t2s` builds the directories and stubs. It will not overwrite
what you already have.

```bash
# copy a tree from Claude / ChatGPT, then:
cd ~/src/new-app
t2s -d          # preview
t2s             # write (asks first)
```

macOS with an empty argv reads the clipboard (`pbpaste`). Everywhere
else, pipe it:

```
myapp/
├── cmd/
│   └── main.go
└── pkg/
    └── util.go
```

```bash
pbpaste | t2s -root ./myapp
```

Go files get a real package stub (`package main` + `func main()` when
the name is `main.go`). Other extensions get a one-line comment header.

## Install

```bash
# Homebrew
brew install lancekrogers/tap/tree2scaffold

# npm (prebuilt `tree2scaffold` and `t2s`)
npm install -g tree2scaffold

# from this tree
make install            # /usr/local/bin + t2s alias
make install PREFIX="$HOME/.local"
```

## Usage

```bash
t2s -root ./myproject   # clipboard or stdin
t2s -d                  # dry-run, then confirm
t2s -yes                # no prompt (scripts)
```

| Flag | What |
|------|------|
| `-root <path>` | Where to write (default `.`) |
| `-d`, `-dry-run` | Show the plan, do not write |
| `-yes` | Skip the confirm prompt |
| `-force` | Overwrite file/dir conflicts |
| `-debug` | Extra logging |

Accepted input: `tree` output (`├──` / `└──`), indented trees with
trailing slashes, or one path per line.

## Why this exists

Agents love drawing project trees and then watching you mkdir by hand.
`t2s` is the missing last step. Never clobbers existing files. Verifies
the tree after it writes.

## WASI

```bash
just build wasm
printf 'myproj/\n├── cmd/\n│   └── main.go\n' \
  | wasmtime run --dir .::/ --env PWD=/ bin/tree2scaffold.wasm -- -root . -yes
```

No clipboard under WASI. Pipe stdin. See the comments in
`justfile` for wazero embedding.

## Custom generators

```go
generator := scaffold.NewDefaultContentGenerator()
generator.RegisterGenerator(".py", func(path, comment string) string {
    return "def main():\n    pass\n"
})
```

`ContentGenerator` is the interface if you want your own.

## Tests

```bash
go test ./...
make integration
make test-all
```

## License

[MIT](LICENSE)
