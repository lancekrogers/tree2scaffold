package parser

import (
	"bufio"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

// Match both tree format lines and simple file list lines
var lineRe = regexp.MustCompile(`^[\s│├└─]*([\w\.\-\/]+)\s*(?:#\s*(.+))?$`)
var simpleFileRe = regexp.MustCompile(`^([\w\.\-\/]+)\s*(?:#\s*(.+))?$`)

type Node struct {
	Path    string // e.g. "cmd/tree2scaffold/main.go" or "pkg/parser/"
	IsDir   bool
	Comment string
}

// Parse reads an ASCII-tree from r and returns Nodes with full relative paths.
// It ignores the very first top-level directory and any lines without a valid name.
// It now supports: 
// - tree format (with full tree starting with root directory)
// - simple file lists (without tree characters)
// - partial tree output (starting with a file like ├── orchestrator.go)
func Parse(r io.Reader) ([]Node, error) {
	scanner := bufio.NewScanner(r)
	var raw []Node
	var parents []string
	first := true
	var rootName string
	
	// Detect format: simple file list or partial tree
	isSimpleFormat := false
	var buffer []string
	
	// Read the first few lines to check format
	for i := 0; i < 3 && scanner.Scan(); i++ {
		buffer = append(buffer, scanner.Text())
		line := buffer[i]
		
		// Check if it's a simple format (no tree chars)
		if !containsTreeChar(line) && simpleFileRe.MatchString(line) {
			isSimpleFormat = true
		}
		
		// Check if it's a partial tree starting with a file
		if i == 0 && containsTreeChar(line) && strings.Contains(line, "├──") {
			fileMatch := lineRe.FindStringSubmatch(line)
			if fileMatch != nil && !strings.HasSuffix(fileMatch[1], "/") {
				// This is a partial tree starting with a file
				first = false // Skip the "first directory is root" logic
			}
		}
	}
	
	// Process buffer and continue scanning
	for i, line := range buffer {
		processLine(line, &raw, &parents, &first, &rootName, isSimpleFormat)
		if i == len(buffer)-1 && scanner.Scan() {
			processLine(scanner.Text(), &raw, &parents, &first, &rootName, isSimpleFormat)
		}
	}
	
	// Continue with the rest of the input
	for scanner.Scan() {
		processLine(scanner.Text(), &raw, &parents, &first, &rootName, isSimpleFormat)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return raw, nil
}

// processLine handles parsing a single line of input
func processLine(line string, raw *[]Node, parents *[]string, first *bool, rootName *string, isSimpleFormat bool) {
	var m []string
	
	if isSimpleFormat {
		m = simpleFileRe.FindStringSubmatch(line)
	} else {
		m = lineRe.FindStringSubmatch(line)
	}
	
	if m == nil {
		return // skip blank or drawing-only lines
	}

	name := m[1]
	comment := ""
	if len(m) > 2 {
		comment = strings.TrimSpace(m[2])
	}
	isDir := strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")

	var depth int
	if isSimpleFormat {
		// For simple format, all files belong to the root directory
		depth = 0
	} else {
		depth = calcDepth(line)
	}

	// The very first directory is the root; capture and skip it
	// Skip this logic if we've already determined this is a partial tree
	if *first {
		*first = false
		if isDir {
			*rootName = name + "/"
			return
		}
	}

	// Ensure parents slice is large enough
	for depth >= len(*parents) {
		*parents = append(*parents, "")
	}
	*parents = (*parents)[:depth+1]
	(*parents)[depth] = name

	// Build full path relative to root
	full := filepath.Join((*parents)[:depth+1]...)
	if isDir {
		full += "/"
	}

	// Strip leading rootName/ if present
	if *rootName != "" && strings.HasPrefix(full, *rootName) {
		full = strings.TrimPrefix(full, *rootName)
	}

	*raw = append(*raw, Node{
		Path:    full,
		IsDir:   isDir,
		Comment: comment,
	})
}

// calcDepth counts how many "tree" characters precede the name and divides by 4.
func calcDepth(line string) int {
	count := 0
	for _, r := range line {
		if strings.ContainsRune(" \t│├└─", r) {
			count++
			continue
		}
		break
	}
	return count / 4
}

// containsTreeChar checks if a line contains ASCII tree characters
func containsTreeChar(line string) bool {
	return strings.ContainsAny(line, "│├└─")
}