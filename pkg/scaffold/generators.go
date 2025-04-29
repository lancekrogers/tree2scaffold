package scaffold

import (
   "fmt"
   "path/filepath"
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
   if comment != "" {
       return fmt.Sprintf("// %s\n\npackage %s\n\nfunc main() {\n    // TODO: implement %s\n}\n", comment, pkg, name)
   }
   return fmt.Sprintf("package %s\n\nfunc main() {\n    // TODO: implement %s\n}\n", pkg, name)
}
