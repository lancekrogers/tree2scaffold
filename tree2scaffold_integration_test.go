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

	// Test case for simple file list format
	t.Run("simple file list format", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Get the directory name to check package later
		rootDirName := filepath.Base(rootDir)

		// Extract the root dir name for package name comparison later
		expectedPackage := strings.ToLower(rootDirName)
		expectedPackage = strings.ReplaceAll(expectedPackage, "-", "_")
		expectedPackage = strings.ReplaceAll(expectedPackage, ".", "_")
		if strings.HasPrefix(expectedPackage, "test_") {
			expectedPackage = strings.TrimPrefix(expectedPackage, "test_")
		}

		// Simple list format input
		input := `orchestrator/
orchestrator.go # Entry point: bootstraps guild, agents, etc.
runner.go # Manages execution loop / agent task polling
dispatcher.go # Handles assigning tasks, managing blocked queue
eventbus.go # Connects to ZeroMQ, publishes/subscribes
`

		// Run the scaffold tool
		proc := exec.Command(exePath, "-root", rootDir)
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Expected files
		expected := []string{
			"orchestrator.go",
			"runner.go",
			"dispatcher.go",
			"eventbus.go",
		}

		// Verify expected files exist
		for _, rel := range expected {
			fullPath := filepath.Join(rootDir, rel)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected file %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking file %s: %v", rel, err)
			}
		}

		// Verify that comments are included in the files
		commentMap := map[string]string{
			"orchestrator.go": "Entry point: bootstraps guild, agents, etc.",
			"runner.go":       "Manages execution loop / agent task polling",
			"dispatcher.go":   "Handles assigning tasks, managing blocked queue",
			"eventbus.go":     "Connects to ZeroMQ, publishes/subscribes",
		}

		for file, expectedComment := range commentMap {
			fullPath := filepath.Join(rootDir, file)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("failed to read %s: %v", file, err)
				continue
			}

			content := string(data)
			// Check that comment is included
			if !strings.Contains(content, "// "+expectedComment) {
				t.Errorf("%s: missing comment %q in file contents", file, expectedComment)
			}

			// Check that the package name matches the directory name
			expectedPackageDecl := "package " + expectedPackage
			if !strings.Contains(content, expectedPackageDecl) {
				t.Errorf("%s: incorrect package name, expected %q in file contents: %s",
					file, expectedPackageDecl, content)
			}
		}
	})

	// Test case for partial tree format (starting directly with a file)
	t.Run("partial tree format", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Get the directory name to check package later
		rootDirName := filepath.Base(rootDir)

		// Extract the root dir name for package name comparison later
		expectedPackage := strings.ToLower(rootDirName)
		expectedPackage = strings.ReplaceAll(expectedPackage, "-", "_")
		expectedPackage = strings.ReplaceAll(expectedPackage, ".", "_")
		if strings.HasPrefix(expectedPackage, "test_") {
			expectedPackage = strings.TrimPrefix(expectedPackage, "test_")
		}

		// Partial tree format input (copy-pasted from tree command output)
		input := `├── orchestrator.go # Entry point for the application
├── runner.go # Handles the execution pipeline
├── dispatcher.go # Routes tasks to workers
└── eventbus.go # Manages pub/sub events
`

		// Run the scaffold tool
		proc := exec.Command(exePath, "-root", rootDir)
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Expected files
		expected := []string{
			"orchestrator.go",
			"runner.go",
			"dispatcher.go",
			"eventbus.go",
		}

		// Verify expected files exist
		for _, rel := range expected {
			fullPath := filepath.Join(rootDir, rel)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected file %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking file %s: %v", rel, err)
			}
		}

		// Verify that comments are included in the files
		commentMap := map[string]string{
			"orchestrator.go": "Entry point for the application",
			"runner.go":       "Handles the execution pipeline",
			"dispatcher.go":   "Routes tasks to workers",
			"eventbus.go":     "Manages pub/sub events",
		}

		for file, expectedComment := range commentMap {
			fullPath := filepath.Join(rootDir, file)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("failed to read %s: %v", file, err)
				continue
			}

			content := string(data)
			// Check that comment is included
			if !strings.Contains(content, "// "+expectedComment) {
				t.Errorf("%s: missing comment %q in file contents", file, expectedComment)
			}

			// Check that the package name matches the directory name
			expectedPackageDecl := "package " + expectedPackage
			if !strings.Contains(content, expectedPackageDecl) {
				t.Errorf("%s: incorrect package name, expected %q in file contents: %s",
					file, expectedPackageDecl, content)
			}
		}
	})
}

