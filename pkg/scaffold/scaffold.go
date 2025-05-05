package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
)

// Validate performs a dry-run check to see if the scaffold operation would succeed
func Validate(root string, nodes []parser.Node) error {
	// First generate all directory paths that will need to be created
	paths := make(map[string]bool) // path -> isDir
	
	// Mark all explicit directories
	for _, n := range nodes {
		if n.IsDir {
			paths[n.Path] = true
		}
	}
	
	// Mark all parent directories of files
	for _, n := range nodes {
		if !n.IsDir {
			// Extract all parent directories
			dir := filepath.Dir(n.Path)
			for dir != "." {
				paths[dir] = true
				dir = filepath.Dir(dir)
			}
		}
	}
	
	// Check for files that would need to be converted to directories
	for dir := range paths {
		dirPath := filepath.Join(root, dir)
		
		// Check if the path exists but is a file
		fileInfo, err := os.Stat(dirPath)
		if err == nil && !fileInfo.IsDir() {
			return fmt.Errorf("cannot create directory %s: a file with the same name already exists", dirPath)
		}
	}
	
	return nil
}

// ForceMode controls whether to overwrite existing files
var ForceMode bool = false

// Apply walks nodes, creating directories and files under root.
func Apply(root string, nodes []parser.Node, onCreate func(path string, isDir bool)) error {
	var stack []parser.Node
	// Get the base name of the root directory to use as package name for top-level files
	rootDirName := filepath.Base(root)
	
	// Process nodes in two phases: first directories, then files
	// First: Create a map to deduplicate paths and identify directories
	paths := make(map[string]bool) // path -> isDir
	
	// Mark all explicit directories
	for _, n := range nodes {
		if n.IsDir {
			paths[n.Path] = true
		}
	}
	
	// Mark all parent directories of files
	for _, n := range nodes {
		if !n.IsDir {
			// Extract all parent directories
			dir := filepath.Dir(n.Path)
			for dir != "." {
				paths[dir] = true
				dir = filepath.Dir(dir)
			}
		}
	}
	
	// First create all directories
	for dir, isDir := range paths {
		if isDir {
			dirPath := filepath.Join(root, dir)
			
			// Special handling for hidden directories which often exist as files first
			isHidden := len(dir) > 0 && dir[0] == '.'
			
			// Check if path exists and is a file
			fileInfo, err := os.Stat(dirPath)
			if err == nil && !fileInfo.IsDir() {
				// Path exists but is a file - remove it before creating directory
				if err := os.Remove(dirPath); err != nil {
					if ForceMode {
						// In force mode, try more aggressively to remove the file
						if removeErr := os.RemoveAll(dirPath); removeErr != nil {
							return fmt.Errorf("cannot convert file to directory even in force mode: %s: %w", dirPath, removeErr)
						}
						// For hidden directories, we log this as it's a common source of issues
						if isHidden {
							fmt.Fprintf(os.Stderr, "Note: Force converted file to directory: %s\n", dirPath)
						}
					} else {
						return fmt.Errorf("cannot convert file to directory: %s: %w", dirPath, err)
					}
				} else {
					// Successfully removed the file
					// For hidden directories, we log this as it's a common source of issues
					if isHidden {
						fmt.Fprintf(os.Stderr, "Note: Converting file to directory: %s\n", dirPath)
					}
				}
			}
			
			if onCreate != nil {
				onCreate(dirPath, true)
			}
			
			// Create the directory
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				return err
			}
		}
	}
	
	// Now process file nodes
	for _, n := range nodes {
		if n.IsDir {
			stack = append(stack, n)
			continue
		}
		
		full := filepath.Join(root, n.Path)
		
		// Check if the path exists and handle conflicts
		fileInfo, err := os.Stat(full)
		if err == nil {
			// Path exists, check if it's already the correct type
			existingIsDir := fileInfo.IsDir()
			if existingIsDir && !n.IsDir {
				// It's a directory but we want to create a file
				// This is a conflict, better skip it
				continue
			} else if !existingIsDir && n.IsDir {
				// It's a file but we want to create a directory
				// Try to remove the file before creating the directory
				if err := os.Remove(full); err != nil {
					// If we can't remove the file, skip this node
					continue
				}
			} else if existingIsDir && n.IsDir {
				// It's already a directory, no need to create it again
				if onCreate != nil {
					onCreate(full, true)
				}
				continue
			}
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