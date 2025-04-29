package scaffold

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
)

// FileGenerator produces the initial content for a file at relPath, given its comment.
type FileGenerator func(relPath, comment string) string

// generators maps extensions to their FileGenerator.
var generators = map[string]FileGenerator{}

// RegisterGenerator associates an extension (e.g. ".go") with its generator.
func RegisterGenerator(ext string, gen FileGenerator) {
	generators[ext] = gen
}

func init() {
	// Re-add Go generator so .go files get package stubs
	RegisterGenerator(".go", generateGo)
	// all other extensions will fall back to defaultGenerator
}

// commentSyntax maps extensions to their line-comment markers.
var commentSyntax = map[string]struct{ prefix, suffix string }{
	".py":   {"# ", ""},
	".js":   {"// ", ""},
	".ts":   {"// ", ""},
	".rs":   {"// ", ""},
	".java": {"// ", ""},
	".c":    {"// ", ""},
	".cpp":  {"// ", ""},
	".h":    {"// ", ""},
	".sh":   {"# ", ""},
	".yaml": {"# ", ""},
	".yml":  {"# ", ""},
	".toml": {"# ", ""},
	".xml":  {"<!-- ", " -->"},
	".html": {"<!-- ", " -->"},
	".md":   {"<!-- ", " -->"},
}

// defaultGenerator emits only the comment header in the right syntax.
func defaultGenerator(relPath, comment string) string {
	if comment == "" {
		return ""
	}
	ext := filepath.Ext(relPath)
	syn, ok := commentSyntax[ext]
	if !ok {
		syn = commentSyntax[".sh"] // fallback
	}
	if syn.suffix != "" {
		return fmt.Sprintf("%s%s%s\n", syn.prefix, comment, syn.suffix)
	}
	return fmt.Sprintf("%s%s\n", syn.prefix, comment)
}

// generateGo produces the package stub for .go files.
func generateGo(relPath, comment string) string {
	pkg := inferPkg(relPath)
	name := filepath.Base(relPath)
	var buf bytes.Buffer
	// raw-string literal for the template
	const tpl = `
{{- if .Comment}}// {{.Comment}}

{{- end}}package {{.Pkg}}

func main() {
    // TODO: implement {{.File}}
}
`
	t := template.Must(template.New("go").Parse(tpl))
	_ = t.Execute(&buf, map[string]string{
		"Pkg":     pkg,
		"File":    name,
		"Comment": comment,
	})
	return buf.String()
}
