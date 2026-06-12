package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestWASMScaffold builds (or reuses) the wasip1 artifact and runs it under a
// WASI runtime to verify it scaffolds files onto the host filesystem through a
// preopened directory. It is skipped unless TEST_ALL/CI is set AND a runtime is
// available, so the default `go test` stays green on machines without one.
//
// A prebuilt module can be supplied via T2S_WASM; the runtime can be forced via
// T2S_WASM_RUNTIME (otherwise wasmtime, then wasmer, are tried).
func TestWASMScaffold(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("TEST_ALL") == "" {
		t.Skip("Skipping WASM integration test in non-CI environment. Set TEST_ALL=1 to run all tests.")
	}

	runtime := findWASMRuntime()
	if runtime == "" {
		t.Skip("no WASI runtime found (set T2S_WASM_RUNTIME, or install wasmtime/wasmer)")
	}

	wasmPath := os.Getenv("T2S_WASM")
	if wasmPath == "" {
		wasmPath = buildWASM(t)
	} else if _, err := os.Stat(wasmPath); err != nil {
		t.Fatalf("T2S_WASM=%s not found: %v", wasmPath, err)
	}

	root := t.TempDir()
	spec := "myproj/\n" +
		"├── cmd/\n" +
		"│   └── main.go\n" +
		"├── pkg/\n" +
		"│   └── util.py # helper\n" +
		"└── go.mod\n"

	cmd := exec.Command(runtime, wasmRuntimeArgs(runtime, wasmPath, root)...)
	cmd.Stdin = bytes.NewBufferString(spec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running wasm under %s failed: %v\n%s", filepath.Base(runtime), err, out)
	}

	// The first top-level directory (myproj/) is stripped by the parser.
	for _, rel := range []string{"cmd/main.go", "pkg/util.py", "go.mod"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("expected %s to be created: %v\noutput:\n%s", rel, err, out)
		}
	}

	// exec-based probes are inert under WASI, so the version falls back to the
	// built-in default and the module name to the guarded default; both lines
	// must still be present and well-formed.
	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("reading generated go.mod: %v", err)
	}
	if !bytes.Contains(goMod, []byte("module ")) || !bytes.Contains(goMod, []byte("go ")) {
		t.Errorf("generated go.mod missing module/go lines:\n%s", goMod)
	}
}

// findWASMRuntime resolves a WASI runtime path, honoring T2S_WASM_RUNTIME first.
func findWASMRuntime() string {
	if rt := os.Getenv("T2S_WASM_RUNTIME"); rt != "" {
		if p, err := exec.LookPath(rt); err == nil {
			return p
		}
		return ""
	}
	for _, name := range []string{"wasmtime", "wasmer"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// wasmRuntimeArgs builds the runtime invocation that preopens mountDir as the
// guest root, sets PWD=/, and forwards the CLI flags to the module.
func wasmRuntimeArgs(runtime, wasmPath, mountDir string) []string {
	switch filepath.Base(runtime) {
	case "wasmer":
		return []string{"run", "--mapdir", "/:" + mountDir, "--env", "PWD=/", wasmPath, "--", "-root", ".", "-yes"}
	default: // wasmtime
		return []string{"run", "--dir", mountDir + "::/", "--env", "PWD=/", wasmPath, "--", "-root", ".", "-yes"}
	}
}

// buildWASM compiles the wasip1 artifact into a temp file and returns its path.
func buildWASM(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "tree2scaffold.wasm")
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", out, "../cmd/tree2scaffold")
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building wasip1 binary: %v\n%s", err, b)
	}
	return out
}
