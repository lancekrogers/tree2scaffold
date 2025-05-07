// Package scaffold provides functionality to convert parsed tree structures into actual file system artifacts.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
)

// Scaffolder is the interface for creating file system structures from parsed tree nodes
type Scaffolder interface {
	// Validate checks if the scaffolding operation would succeed
	Validate(root string, nodes []parser.Node) error
	
	// Apply creates the directory and file structure on disk
	Apply(root string, nodes []parser.Node, callback CreationCallback) error
	
	// VerifyStructure checks if the created structure matches the specification
	VerifyStructure(root string, nodes []parser.Node) error
}

// CreationCallback is called when a file or directory is created
type CreationCallback func(path string, isDir bool)

// ContentGenerator generates content for files
type ContentGenerator interface {
	// GenerateContent creates content for a file based on its path and comment
	GenerateContent(relPath string, comment string) string
	
	// RegisterGenerator adds a new generator for a specific extension or filename
	RegisterGenerator(extOrName string, generator FileGenerator)
}

// DefaultScaffolder implements the Scaffolder interface with default behavior
type DefaultScaffolder struct {
	ForceMode       bool
	ContentProvider ContentGenerator
}

// NewScaffolder creates a new default scaffolder
func NewScaffolder() *DefaultScaffolder {
	return &DefaultScaffolder{
		ForceMode:       false,
		ContentProvider: NewDefaultContentGenerator(),
	}
}

// NewScaffolderWithForce creates a new scaffolder with force mode enabled
func NewScaffolderWithForce() *DefaultScaffolder {
	return &DefaultScaffolder{
		ForceMode:       true,
		ContentProvider: NewDefaultContentGenerator(),
	}
}

// ForceMode controls whether to overwrite existing files (backward compatibility)
var ForceMode bool = false

// Validate performs a dry-run check to see if the scaffold operation would succeed
func (s *DefaultScaffolder) Validate(root string, nodes []parser.Node) error {
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

// VerifyStructure ensures the directory structure matches the specification after creation
func (s *DefaultScaffolder) VerifyStructure(root string, nodes []parser.Node) error {
	// Map of all expected paths
	expectedPaths := make(map[string]bool)
	
	// Add all files and directories to expected paths
	for _, n := range nodes {
		expectedPaths[n.Path] = true
	}
	
	// Use a file system walker to verify all expected paths exist
	missingPaths := []string{}
	
	// Check each expected path
	for path := range expectedPaths {
		fullPath := filepath.Join(root, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missingPaths = append(missingPaths, path)
		}
	}
	
	// If any paths are missing, report the error
	if len(missingPaths) > 0 {
		return fmt.Errorf("structure verification failed: missing %d paths including %v", 
			len(missingPaths), missingPaths[:min(3, len(missingPaths))])
	}
	
	return nil
}

// Apply walks nodes, creating directories and files under root.
func (s *DefaultScaffolder) Apply(root string, nodes []parser.Node, onCreate CreationCallback) error {
	var stack []parser.Node
	// Process nodes in a structured way
	
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
					if s.ForceMode || ForceMode {
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
			} else if !existingIsDir && !n.IsDir {
				// It's a file and we want to create a file
				// Skip - don't overwrite existing files
				fmt.Fprintf(os.Stderr, "Note: Skipping existing file: %s\n", full)
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

		// Generate content using the content provider
		var content string
		fileName := filepath.Base(n.Path)
		
		// Check if file is main.go - special handling for main.go files
		if fileName == "main.go" {
			// main.go files always get package main
			content = generateMainGoFile(n.Path, comment)
		} else {
			// Generate content through the provider
			content = s.ContentProvider.GenerateContent(n.Path, comment)
		}

		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Optional: Verify the scaffolded structure matches the specification
	return s.VerifyStructure(root, nodes)
}

// generateMainGoFile generates content specifically for main.go files
func generateMainGoFile(relPath, comment string) string {
	if comment != "" {
		return fmt.Sprintf("// %s\n\npackage main\n\nfunc main() {\n    // TODO: implement main.go\n}\n", comment)
	}
	return fmt.Sprintf("package main\n\nfunc main() {\n    // TODO: implement main.go\n}\n")
}

// Backward compatibility function to maintain the old API
func Validate(root string, nodes []parser.Node) error {
	s := NewScaffolder()
	return s.Validate(root, nodes)
}

// Backward compatibility function to maintain the old API
func Apply(root string, nodes []parser.Node, onCreate CreationCallback) error {
	s := NewScaffolder()
	s.ForceMode = ForceMode
	return s.Apply(root, nodes, onCreate)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}