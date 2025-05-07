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
	// Skip if not running in CI to avoid interrupting manual tests
	if os.Getenv("CI") == "" && os.Getenv("TEST_ALL") == "" {
		t.Skip("Skipping integration test in non-CI environment. Set TEST_ALL=1 to run all tests.")
	}
	// Build the CLI binary
	buildDir := t.TempDir()
	exePath := filepath.Join(buildDir, "tree2scaffold")
	buildCmd := exec.Command("go", "build", "-o", exePath, "../cmd/tree2scaffold")
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

	// Test case for nested directory tree format
	t.Run("nested directory tree format", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Nested directory tree format input
		input := `algo-scales/
├── cmd/
├── internal/
│   ├── license/
│   ├── api/
│   ├── problem/
│   ├── session/
│   └── ui/
├── server/
├── algo-scales.nvim/
│   └── lua/algo-scales/
`

		// Run the scaffold tool
		proc := exec.Command(exePath, "-root", rootDir)
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Expected directories
		expected := []string{
			"cmd",
			"internal",
			"internal/license",
			"internal/api",
			"internal/problem",
			"internal/session",
			"internal/ui",
			"server",
			"algo-scales.nvim",
			"algo-scales.nvim/lua/algo-scales",
		}

		// Verify expected directories exist
		for _, rel := range expected {
			fullPath := filepath.Join(rootDir, rel)
			if info, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected directory %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking directory %s: %v", rel, err)
			} else if !info.IsDir() {
				t.Errorf("expected %s to be a directory, but it's a file", rel)
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

	// Test case for complex nested tree with special directory handling
	t.Run("complex nested tree", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Complex nested tree with UI files and test_problem.json
		input := `algo-scales/
├── main.go                            # Main entry point for the application
├── go.mod                             # Go module definition and dependencies
├── go.sum                             # Checksums for dependencies
├── Makefile                           # Build automation and commands
├── README.md                          # Project documentation and usage guide
├── TESTING.md                         # Testing guide and instructions
├── docker-compose.yml                 # Docker configuration for API server
├── .github
│   └── workflows
│       └── build.yml                  # GitHub Actions CI/CD workflow
├── internal
│   ├── api
│   │   ├── client.go                  # API client for problem downloads
│   │   └── client_test.go             # Tests for API client
│   ├── ui
│       ├── ui.go                      # Terminal UI using Bubble Tea
│       ├── ui_test.go                 # Tests for UI components
│       └── code.go                    # Syntax highlighting for code display
├── server
│   ├── main.go                        # API server implementation
│   └── Dockerfile                     # Container definition for server
└── testdata
    └── problems
        └── test_problem.json          # Sample problem for testing
`

		// Run the scaffold tool
		proc := exec.Command(exePath, "-root", rootDir)
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Expected directories
		expectedDirs := []string{
			".github",
			".github/workflows",
			"internal",
			"internal/api",
			"internal/ui",
			"server",
			"testdata",
			"testdata/problems",
		}

		// Verify expected directories exist
		for _, rel := range expectedDirs {
			fullPath := filepath.Join(rootDir, rel)
			if info, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected directory %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking directory %s: %v", rel, err)
			} else if !info.IsDir() {
				t.Errorf("expected %s to be a directory, but it's a file", rel)
			}
		}

		// Expected files with special handling
		expectedFiles := []string{
			"main.go",
			"go.mod",
			"go.sum",
			"Makefile",
			"README.md",
			"TESTING.md",
			"docker-compose.yml",
			".github/workflows/build.yml", // The build.yml file should be in workflows directory
			"internal/api/client.go",
			"internal/api/client_test.go",
			"internal/ui/ui.go",
			"internal/ui/ui_test.go",
			"internal/ui/code.go",
			"server/main.go",
			"server/Dockerfile",
			"testdata/problems/test_problem.json",
		}

		// Verify expected files exist in the correct locations
		for _, rel := range expectedFiles {
			fullPath := filepath.Join(rootDir, rel)
			if info, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected file %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking file %s: %v", rel, err)
			} else if info.IsDir() {
				t.Errorf("expected %s to be a file, but it's a directory", rel)
			}
		}
	})

	// Test case for force mode with hidden files
	t.Run("force_mode_with_hidden_files", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Create a conflicting hidden file
		if err := os.WriteFile(filepath.Join(rootDir, ".github"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Input with a directory that would conflict with the hidden file
		input := `myapp/
├── .github
│   └── workflows
│       └── build.yml                  # GitHub Actions CI/CD workflow
`

		// Run the scaffold tool with force flag
		proc := exec.Command(exePath, "-root", rootDir, "-force")
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Verify expected directories exist
		expectedDirs := []string{
			".github",
			".github/workflows",
		}

		// Verify expected directories exist
		for _, rel := range expectedDirs {
			fullPath := filepath.Join(rootDir, rel)
			if info, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected directory %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking directory %s: %v", rel, err)
			} else if !info.IsDir() {
				t.Errorf("expected %s to be a directory, but it's a file", rel)
			}
		}

		// Verify that file was created
		expectedFile := ".github/workflows/build.yml"
		fullPath := filepath.Join(rootDir, expectedFile)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", expectedFile)
		}
	})

	// Test case for multiple hidden directory conventions
	t.Run("multiple_hidden_directory_conventions", func(t *testing.T) {
		// Create a fresh root for scaffolding
		rootDir := t.TempDir()

		// Input with multiple hidden directory structures
		input := `project/
├── .github
│   ├── workflows
│   │   ├── build.yml                  # GitHub Actions CI/CD workflow
│   │   └── release.yml                # Release automation
│   └── settings.yml                   # Repository settings
├── .vscode
│   ├── tasks.json                     # VS Code tasks
│   └── settings.json                  # VS Code settings
`

		// Run the scaffold tool
		proc := exec.Command(exePath, "-root", rootDir)
		proc.Stdin = bytes.NewBufferString(input)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		if err := proc.Run(); err != nil {
			t.Fatalf("tree2scaffold execution failed: %v", err)
		}

		// Verify expected directories exist
		expectedDirs := []string{
			".github",
			".github/workflows",
			".vscode",
		}

		// Verify expected directories exist
		for _, rel := range expectedDirs {
			fullPath := filepath.Join(rootDir, rel)
			if info, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected directory %s not found", rel)
			} else if err != nil {
				t.Errorf("error checking directory %s: %v", rel, err)
			} else if !info.IsDir() {
				t.Errorf("expected %s to be a directory, but it's a file", rel)
			}
		}

		// Verify expected files exist in the correct locations
		expectedFiles := []string{
			".github/workflows/build.yml",
			".github/workflows/release.yml",
			".github/settings.yml",
			".vscode/tasks.json",
			".vscode/settings.json",
		}

		// Verify expected files exist
		for _, rel := range expectedFiles {
			fullPath := filepath.Join(rootDir, rel)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("expected file %s not found", rel)
			}
		}
	})
}
