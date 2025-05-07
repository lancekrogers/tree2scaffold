package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackageInference tests that the package name is correctly inferred from the directory
func TestPackageInference(t *testing.T) {
	// Skip in non-CI environments unless specifically requested
	if os.Getenv("CI") == "" && os.Getenv("TEST_ALL") == "" {
		t.Skip("Skipping test in non-CI environment. Set TEST_ALL=1 to run all tests.")
	}
	
	// Create a temporary directory for our test
	tmpDir := t.TempDir()
	
	// Create specific files in well-known directories
	paths := map[string]string{
		"cmd/regular.go":       "# Command file",
		"cmd/main.go":          "# Main file",
		"internal/util/util.go": "# Utility file",
	}
	
	// Create the files
	for path, comment := range paths {
		fullPath := filepath.Join(tmpDir, path)
		
		// Create directory
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}
		
		// Create a sample tree input for each file individually
		treeInput := fmt.Sprintf("test/\n├── %s    %s", path, comment)
		
		// Run tree2scaffold for each file
		// Make sure we're using the binary from the bin directory 
		binaryPath := filepath.Join("..", "bin", "tree2scaffold")
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			binaryPath = "tree2scaffold" // Fall back to PATH lookup
		}
		
		cmd := exec.Command(binaryPath, "-root", tmpDir, "-yes")
		cmd.Stdin = strings.NewReader(treeInput)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("tree2scaffold failed for %s: %v\nOutput: %s", path, err, string(output))
		}
	}
	
	// Verify the package names
	expectedPackages := map[string]string{
		"cmd/regular.go":       "package cmd",
		"cmd/main.go":          "package main",
		"internal/util/util.go": "package util",
	}
	
	for path, expectedPackage := range expectedPackages {
		content, err := os.ReadFile(filepath.Join(tmpDir, path))
		if err != nil {
			t.Errorf("Failed to read %s: %v", path, err)
			continue
		}
		
		if !strings.Contains(string(content), expectedPackage) {
			t.Errorf("File %s has incorrect package: expected %q, got:\n%s", 
				path, expectedPackage, string(content))
		} else {
			t.Logf("OK: %s correctly has %s", path, expectedPackage)
		}
	}
}