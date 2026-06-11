// Package scaffold provides functionality to convert parsed tree structures into actual file system artifacts.
package scaffold

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/tree2scaffold/internal/env"
)

// FileGenerator produces the initial content for a file at relPath, given its comment.
type FileGenerator func(relPath, comment string) string

// DefaultContentGenerator implements the ContentGenerator interface
type DefaultContentGenerator struct {
	env           env.Environment
	generators    map[string]FileGenerator
	commentSyntax map[string]struct{ prefix, suffix string }
}

// NewDefaultContentGenerator creates a new content generator with default file
// handlers, probing the host environment for the Go toolchain version and VCS
// remote. Under WASI those probes degrade to sensible defaults automatically.
func NewDefaultContentGenerator() *DefaultContentGenerator {
	return newContentGenerator(env.New())
}

// newContentGenerator builds a generator with an explicit environment so tests
// can inject a stub (e.g. to exercise the WASI-degraded fallbacks on native).
func newContentGenerator(e env.Environment) *DefaultContentGenerator {
	gen := &DefaultContentGenerator{
		env:        e,
		generators: make(map[string]FileGenerator),
		commentSyntax: map[string]struct{ prefix, suffix string }{
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
			".mod":  {"// ", ""}, // go.mod files use Go-style comments
			".work": {"// ", ""}, // go.work files use Go-style comments
			".sum":  {"// ", ""}, // go.sum files use Go-style comments
			".go":   {"// ", ""}, // Go files
		},
	}

	// Register default generators
	gen.RegisterGenerator(".go", gen.generateGo)
	gen.RegisterGenerator("go.mod", gen.generateGoMod)
	gen.RegisterGenerator("go.work", gen.generateGoWork)
	gen.RegisterGenerator("go.sum", gen.generateGoSum)

	return gen
}

// RegisterGenerator adds a new generator for a specific extension or filename
func (g *DefaultContentGenerator) RegisterGenerator(extOrName string, generator FileGenerator) {
	g.generators[extOrName] = generator
}

// GenerateContent creates content for a file based on its path and comment
func (g *DefaultContentGenerator) GenerateContent(relPath, comment string) string {
	fileName := filepath.Base(relPath)
	ext := filepath.Ext(relPath)

	// Check for specific filename generator first (e.g., "go.mod")
	if generator, ok := g.generators[fileName]; ok {
		return generator(relPath, comment)
	}

	// Then try extension-based generator (e.g., ".go")
	if generator, ok := g.generators[ext]; ok {
		return generator(relPath, comment)
	}

	// Fall back to default comment generator
	return g.defaultGenerator(relPath, comment)
}

// defaultGenerator emits only the comment header in the right syntax.
func (g *DefaultContentGenerator) defaultGenerator(relPath, comment string) string {
	if comment == "" {
		return ""
	}

	ext := filepath.Ext(relPath)
	syn, ok := g.commentSyntax[ext]
	if !ok {
		syn = g.commentSyntax[".sh"] // fallback to shell-style comments
	}

	if syn.suffix != "" {
		return fmt.Sprintf("%s%s%s\n", syn.prefix, comment, syn.suffix)
	}
	return fmt.Sprintf("%s%s\n", syn.prefix, comment)
}

// generateGo produces the package stub for .go files.
func (g *DefaultContentGenerator) generateGo(relPath, comment string) string {
	pkg := inferPkg(relPath)
	name := filepath.Base(relPath)

	// Check if this is a main.go file - special handling for main.go
	if name == "main.go" {
		if comment != "" {
			return fmt.Sprintf("// %s\n\npackage main\n\nfunc main() {\n    // TODO: implement %s\n}\n", comment, name)
		}
		return fmt.Sprintf("package main\n\nfunc main() {\n    // TODO: implement %s\n}\n", name)
	}

	// Regular .go file handling
	if comment != "" {
		return fmt.Sprintf("// %s\n\npackage %s\n\n// TODO: implement %s\n", comment, pkg, name)
	}
	return fmt.Sprintf("package %s\n\n// TODO: implement %s\n", pkg, name)
}

// generateGoMod creates a go.mod file with the host Go version (falling back to a
// default when the toolchain cannot be probed, e.g. under WASI).
func (g *DefaultContentGenerator) generateGoMod(relPath, comment string) string {
	moduleName := g.inferModuleName(relPath)
	goVersion := g.goVersion()

	if comment != "" {
		return fmt.Sprintf("// %s\n\nmodule %s\n\ngo %s\n", comment, moduleName, goVersion)
	}
	return fmt.Sprintf("module %s\n\ngo %s\n", moduleName, goVersion)
}

// generateGoWork creates a go.work file for a multi-module workspace.
func (g *DefaultContentGenerator) generateGoWork(relPath, comment string) string {
	goVersion := g.goVersion()

	if comment != "" {
		return fmt.Sprintf("// %s\n\ngo %s\n\nuse (\n    // Add your module directories here\n    // .\n)\n", comment, goVersion)
	}
	return fmt.Sprintf("go %s\n\nuse (\n    // Add your module directories here\n    // .\n)\n", goVersion)
}

// generateGoSum creates a placeholder go.sum file.
func (g *DefaultContentGenerator) generateGoSum(relPath, comment string) string {
	if comment != "" {
		return fmt.Sprintf("// %s\n// This file will be automatically populated when dependencies are added to go.mod\n", comment)
	}
	return "// This file will be automatically populated when dependencies are added to go.mod\n"
}

// goVersion returns the host Go major.minor, falling back to a sane default when
// the toolchain cannot be probed (e.g. exec is unavailable under WASI).
func (g *DefaultContentGenerator) goVersion() string {
	const fallback = "1.24"
	if v, err := g.env.GoVersion(); err == nil && v != "" {
		return v
	}
	return fallback
}

// inferPkg derives the Go package name from relPath.
// Files named main.go get package main; otherwise use the name of the parent directory.
func inferPkg(relPath string) string {
	dirPath := filepath.Dir(relPath)
	fileName := filepath.Base(relPath)

	// main.go files should always be package main
	if fileName == "main.go" {
		return "main"
	}

	// top-level files (Dir == ".") get main package
	if dirPath == "." {
		return "main"
	}

	// Use the directory name as the package name
	return filepath.Base(dirPath)
}

// inferModuleName derives a Go module name from the relative path of a go.mod file.
// This is a best-effort guess based on common conventions. The VCS remote and
// working directory are read through the injected environment, so it degrades to
// a default name when those probes are unavailable (e.g. under WASI).
func (g *DefaultContentGenerator) inferModuleName(relPath string) string {
	// Extract the directory where go.mod is located
	dir := filepath.Dir(relPath)

	// If it's in the root, use the current git remote / directory name
	if dir == "." {
		if remoteURL, err := g.env.GitRemoteOriginURL(); err == nil && strings.Contains(remoteURL, "github.com") {
			// Format: https://github.com/username/repo.git or git@github.com:username/repo.git
			urlParts := strings.Split(remoteURL, "/")
			if len(urlParts) >= 2 {
				repoName := strings.TrimSuffix(urlParts[len(urlParts)-1], ".git")
				userName := urlParts[len(urlParts)-2]
				if strings.Contains(userName, ":") {
					userName = strings.Split(userName, ":")[1]
				}
				return fmt.Sprintf("github.com/%s/%s", userName, repoName)
			}
		}

		// Fallback: use the current directory name (os.Getwd works on wasip1 too).
		// Guard against degenerate bases like "/" (WASI guest root) or ".".
		if cwd, err := g.env.Getwd(); err == nil {
			if base := filepath.Base(cwd); base != "" && base != "/" && base != "." {
				return base
			}
		}

		return "example.com/mymodule"
	}

	// For nested modules, use the directory structure
	// This is a simple implementation and might need to be customized
	return "example.com/" + dir
}
