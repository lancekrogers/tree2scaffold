package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTree2ScaffoldIntegration(t *testing.T) {
	// Build the CLI binary
	buildDir := t.TempDir()
	exePath := filepath.Join(buildDir, "tree2scaffold")
	buildCmd := exec.Command("go", "build", "-o", exePath, "./cmd/tree2scaffold")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build tree2scaffold: %v", err)
	}

	// ASCII tree input from `tree` output including comments
	treeInput := `tree2scaffold/
├── cmd/
│   └── tree2scaffold/
│       └── main.go        # entry-point
│
├── pkg/
│   ├── parser/            # parses pasted ASCII trees
│   │   ├── parser.go
│   │   └── parser_test.go
│   │
│   └── scaffold/          # does os.MkdirAll, file templates, writes files
│       ├── scaffold.go
│       └── scaffold_test.go
│
├── go.mod
├── go.sum
├── README.md
└── .gitignore
`

	// Create a fresh root for scaffolding
	rootDir := t.TempDir()

	// Run the scaffold tool with the tree input
	proc := exec.Command(exePath, "-root", rootDir)
	proc.Stdin = bytes.NewBufferString(treeInput)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	if err := proc.Run(); err != nil {
		t.Fatalf("tree2scaffold execution failed: %v", err)
	}

	// Verify Go files have correct package and injected comments
	checks := []struct {
		path        string
		wantPkg     string
		wantComment string
	}{
		{"cmd/tree2scaffold/main.go", "package main", "// entry-point"},
		{"pkg/parser/parser.go", "package parser", "// parses pasted ASCII trees"},
		{"pkg/parser/parser_test.go", "package parser", "// parses pasted ASCII trees"},
		{"pkg/scaffold/scaffold.go", "package scaffold", "// does os.MkdirAll, file templates, writes files"},
		{"pkg/scaffold/scaffold_test.go", "package scaffold", "// does os.MkdirAll, file templates, writes files"},
	}
	for _, tc := range checks {
		full := filepath.Join(rootDir, tc.path)
		data, err := os.ReadFile(full)
		if err != nil {
			t.Errorf("file %s not found: %v", tc.path, err)
			continue
		}
		content := string(data)
		if !strings.Contains(content, tc.wantPkg) {
			t.Errorf("%s: missing package declaration %q", tc.path, tc.wantPkg)
		}
		if !strings.Contains(content, tc.wantComment) {
			t.Errorf("%s: missing injected comment %q", tc.path, tc.wantComment)
		}
	}

	// Verify other project files exist
	others := []string{"go.mod", "go.sum", "README.md", ".gitignore"}
	for _, f := range others {
		if _, err := os.Stat(filepath.Join(rootDir, f)); err != nil {
			t.Errorf("expected %s to exist, got error: %v", f, err)
		}
	}
}
