// tree2scaffold_integration_test.go
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

	// Run the scaffold tool
	proc := exec.Command(exePath, "-root", rootDir)
	proc.Stdin = bytes.NewBufferString(treeInput)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	if err := proc.Run(); err != nil {
		t.Fatalf("tree2scaffold execution failed: %v", err)
	}

	// Expected paths to exist
	expected := []string{
		"cmd/tree2scaffold/main.go",
		"pkg/parser/parser.go",
		"pkg/parser/parser_test.go",
		"pkg/scaffold/scaffold.go",
		"pkg/scaffold/scaffold_test.go",
		"go.mod",
		"go.sum",
		"README.md",
		".gitignore",
	}

	for _, rel := range expected {
		fullPath := filepath.Join(rootDir, rel)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", rel)
		} else if err != nil {
			t.Errorf("error checking file %s: %v", rel, err)
		}
	}

	// Verify Go files contain correct package declarations
	pkgChecks := []struct {
		path    string
		wantPkg string
	}{
		{"cmd/tree2scaffold/main.go", "package main"},
		{"pkg/parser/parser.go", "package parser"},
		{"pkg/parser/parser_test.go", "package parser"},
		{"pkg/scaffold/scaffold.go", "package scaffold"},
		{"pkg/scaffold/scaffold_test.go", "package scaffold"},
	}
	for _, pc := range pkgChecks {
		fullPath := filepath.Join(rootDir, pc.path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read %s: %v", pc.path, err)
			continue
		}
		content := string(data)
		if !strings.Contains(content, pc.wantPkg) {
			t.Errorf("%s: missing %q in file contents", pc.path, pc.wantPkg)
		}
	}
}
