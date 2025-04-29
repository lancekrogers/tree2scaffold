package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
	"github.com/lancekrogers/tree2scaffold/pkg/scaffold"
)

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
			if err := scaffold.Apply(root, tt.nodes); err != nil {
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
