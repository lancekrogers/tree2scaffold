package parser

import (
	"bufio"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

var lineRe = regexp.MustCompile(`^[\s│├└─]*([\w\.\-\/]+)\s*(?:#\s*(.+))?$`)

type Node struct {
	Path    string // e.g. "cmd/tree2scaffold/main.go" or "pkg/parser/"
	IsDir   bool
	Comment string
}

// Parse reads an ASCII-tree from r and returns Nodes with full relative paths.
// It ignores the very first top-level directory and any lines without a valid name.
func Parse(r io.Reader) ([]Node, error) {
	scanner := bufio.NewScanner(r)
	var raw []Node
	var parents []string
	first := true
	var rootName string

	for scanner.Scan() {
		line := scanner.Text()
		m := lineRe.FindStringSubmatch(line)
		if m == nil {
			continue // skip blank or drawing-only lines
		}

		name := m[1]
		comment := ""
		if len(m) > 2 {
			comment = strings.TrimSpace(m[2])
		}
		isDir := strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")

		depth := calcDepth(line)

		// The very first directory is the root; capture and skip it
		if first {
			first = false
			if isDir {
				rootName = name + "/"
				continue
			}
		}

		// Ensure parents slice is large enough
		for depth >= len(parents) {
			parents = append(parents, "")
		}
		parents = parents[:depth+1]
		parents[depth] = name

		// Build full path relative to root
		full := filepath.Join(parents[:depth+1]...)
		if isDir {
			full += "/"
		}

		// Strip leading rootName/ if present
		if rootName != "" && strings.HasPrefix(full, rootName) {
			full = strings.TrimPrefix(full, rootName)
		}

		raw = append(raw, Node{
			Path:    full,
			IsDir:   isDir,
			Comment: comment,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return raw, nil
}

// calcDepth counts how many “tree” characters precede the name and divides by 4.
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
