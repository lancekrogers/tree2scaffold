package scaffold

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
)

// Apply walks nodes, creating directories and files under root.
func Apply(root string, nodes []parser.Node, onCreate func(path string, isDir bool)) error {
	var stack []parser.Node

	for _, n := range nodes {
		full := filepath.Join(root, n.Path)

		if n.IsDir {
			stack = append(stack, n)
			if onCreate != nil {
				onCreate(full, true)
			}
			if err := os.MkdirAll(full, 0o755); err != nil {
				return err
			}
			continue
		}

		// Determine which comment to use
		comment := n.Comment
		if comment == "" {
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].Comment != "" {
					comment = stack[i].Comment
					break
				}
			}
		}

		if onCreate != nil {
			onCreate(full, false)
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}

		// Choose generator based solely on file extension
		var content string
		switch filepath.Ext(n.Path) {
		case ".go":
			content = generateGo(n.Path, comment)
		default:
			content = defaultGenerator(n.Path, comment)
		}

		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}

// inferPkg derives the Go package name from relPath.
// Files under cmd/ or at the project root get package main;
// otherwise use the name of the parent directory.
func inferPkg(relPath string) string {
   dirPath := filepath.Dir(relPath)
   // top-level files (Dir == ".") or cmd/* are main packages
   if strings.HasPrefix(relPath, "cmd/") || dirPath == "." {
       return "main"
   }
   return filepath.Base(dirPath)
}
