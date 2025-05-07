package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
	"github.com/lancekrogers/tree2scaffold/pkg/scaffold"
)

func TestValidate(t *testing.T) {
	// We'll create temporary directories for each test case

	// Test cases
	testCases := []struct {
		name          string
		setupFunc     func(dir string) error
		nodes         []parser.Node
		expectError   bool
		errorContains string
	}{
		{
			name: "No conflicts",
			setupFunc: func(dir string) error {
				return nil
			},
			nodes: []parser.Node{
				{Path: "dir/", IsDir: true},
				{Path: "dir/file.txt", IsDir: false},
			},
			expectError: false,
		},
		{
			name: "File conflicts with directory",
			setupFunc: func(dir string) error {
				// Create a file that conflicts with a directory
				return os.WriteFile(filepath.Join(dir, "dir"), []byte("test"), 0644)
			},
			nodes: []parser.Node{
				{Path: "dir/", IsDir: true},
				{Path: "dir/file.txt", IsDir: false},
			},
			expectError:   true,
			errorContains: "cannot create directory",
		},
		{
			name: "Hidden file conflicts with directory",
			setupFunc: func(dir string) error {
				// Create a hidden file that conflicts with a directory
				return os.WriteFile(filepath.Join(dir, ".hidden"), []byte("test"), 0644)
			},
			nodes: []parser.Node{
				{Path: ".hidden/", IsDir: true},
				{Path: ".hidden/file.txt", IsDir: false},
			},
			expectError:   true,
			errorContains: "cannot create directory",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new temp dir for each test
			testDir := t.TempDir()

			// Set up test environment
			if err := tc.setupFunc(testDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Create scaffolder and run validation
			s := scaffold.NewScaffolder()
			err := s.Validate(testDir, tc.nodes)

			// Check if error is as expected
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message doesn't contain expected text. Got: %v, Want: %v", err, tc.errorContains)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []parser.Node
		wantFiles map[string]func(string) bool // filepath → validator on contents
	}{
		{
			name: "simple dir+file with comment",
			nodes: []parser.Node{
				{Path: "svc/", IsDir: true, Comment: "service code"},
				{Path: "svc/api.go", IsDir: false, Comment: ""},
			},
			wantFiles: map[string]func(string) bool{
				"svc/api.go": func(c string) bool {
					return strings.Contains(c, "// service code") &&
						strings.Contains(c, "package svc")
				},
			},
		},
		{
			name: "cmd tool → package main",
			nodes: []parser.Node{
				{Path: "cmd/tool/", IsDir: true, Comment: ""},
				{Path: "cmd/tool/main.go", IsDir: false, Comment: ""},
			},
			wantFiles: map[string]func(string) bool{
				"cmd/tool/main.go": func(c string) bool {
					return strings.Contains(c, "package main")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			// Create scaffolder instance
			s := scaffold.NewScaffolder()
			// Pass nil for the onCreate callback
			if err := s.Apply(root, tt.nodes, nil); err != nil {
				t.Fatalf("Apply() error = %v", err)
			}

			for rel, ok := range tt.wantFiles {
				path := filepath.Join(root, rel)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("expected file %s: %v", rel, err)
				}
				content := string(data)
				if !ok(content) {
					t.Errorf("file %s content validation failed:\n%s", rel, content)
				}
			}
		})
	}
}