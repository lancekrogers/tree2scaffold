package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var lineRe = regexp.MustCompile(`^[\s│├└─]*([^\s#][^#]*?)(?:\s*#\s*(.+))?$`)

type Node struct {
	Path    string
	IsDir   bool
	Comment string
}

// Parse reads an ASCII tree from r and returns a slice of Nodes.
func Parse(r io.Reader) ([]Node, error) {
	var out []Node
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m := lineRe.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		rel := strings.TrimSpace(m[1])
		comment := ""
		if len(m) > 2 {
			comment = strings.TrimSpace(m[2])
		}
		isDir := strings.HasSuffix(rel, "/")
		rel = strings.TrimSuffix(rel, "/")
		out = append(out, Node{Path: rel, IsDir: isDir, Comment: comment})
	}
	return out, scanner.Err()
}
