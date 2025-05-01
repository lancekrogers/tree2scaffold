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
	// Get the base name of the root directory to use as package name for top-level files
	rootDirName := filepath.Base(root)

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

		// Choose generator based on file name or extension
		var content string
		fileName := filepath.Base(n.Path)
		ext := filepath.Ext(n.Path)
		
		// First check if we have a specific generator for this filename
		if generator, ok := generators[fileName]; ok {
			content = generator(n.Path, comment)
		} else if generator, ok := generators[ext]; ok {
			// Then try extension-based generators with the root directory name context
			if ext == ".go" && filepath.Dir(n.Path) == "." {
				// For top-level .go files, we'll pass the root directory name to be used
				content = generateGoWithRootPackage(n.Path, comment, rootDirName)
			} else {
				content = generator(n.Path, comment)
			}
		} else {
			// Fall back to default generator
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