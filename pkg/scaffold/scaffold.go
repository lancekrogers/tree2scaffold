package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
)

// Apply creates dirs and files under root per the nodes, injecting comments and package names.
func Apply(root string, nodes []parser.Node) error {
	stack := []parser.Node{}
	for _, n := range nodes {
		full := filepath.Join(root, n.Path)

		if n.IsDir {
			stack = append(stack, n)
			if err := os.MkdirAll(full, 0755); err != nil {
				return err
			}
			continue
		}

		// use file-level comment if present, otherwise fallback to nearest ancestor
		comment := n.Comment
		if comment == "" {
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].Comment != "" {
					comment = stack[i].Comment
					break
				}
			}
		}
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}

		pkg := inferPkg(n.Path)
		content := renderFile(pkg, filepath.Base(n.Path), comment)
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func inferPkg(rel string) string {
	dir := filepath.Base(filepath.Dir(rel))
	if strings.HasPrefix(rel, "cmd/") || dir == "" {
		return "main"
	}
	return dir
}

func renderFile(pkg, file, comment string) string {
	const tmpl = `{{if .Comment}}// {{.Comment}}

{{end}}package {{.Pkg}}

func main() {
    // TODO: implement {{.File}}
}
`
	t := template.Must(template.New("file").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{
		"Pkg":     pkg,
		"File":    file,
		"Comment": comment,
	}); err != nil {
		panic(err)
	}
	return buf.String()
}
