// Package scaffold provides functionality to convert parsed tree structures into actual file system artifacts.
package scaffold

import (
   "fmt"
   "os/exec"
   "path/filepath"
   "strings"
)

// FileGenerator produces the initial content for a file at relPath, given its comment.
type FileGenerator func(relPath, comment string) string

// DefaultContentGenerator implements the ContentGenerator interface
type DefaultContentGenerator struct {
	generators     map[string]FileGenerator
	commentSyntax  map[string]struct{ prefix, suffix string }
}

// NewDefaultContentGenerator creates a new content generator with default file handlers
func NewDefaultContentGenerator() *DefaultContentGenerator {
	gen := &DefaultContentGenerator{
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

// generateGoMod creates a go.mod file with the current Go version.
func (g *DefaultContentGenerator) generateGoMod(relPath, comment string) string {
   // Determine module name based on directory structure
   moduleName := inferModuleName(relPath)
   // Using Go 1.24 as the default version
   goVersion := "1.24"
   
   // Try to get the actual Go version from the environment
   output, err := exec.Command("go", "version").Output()
   if err == nil {
       // Parse version from output like "go version go1.24.2 darwin/arm64"
       versionStr := string(output)
       versionParts := strings.Fields(versionStr)
       if len(versionParts) >= 3 {
           // Extract version number without "go" prefix
           versionFull := strings.TrimPrefix(versionParts[2], "go")
           // Take only major.minor (1.24 from 1.24.2)
           if dotIdx := strings.LastIndex(versionFull, "."); dotIdx > 0 {
               goVersion = versionFull[:dotIdx]
           } else {
               goVersion = versionFull
           }
       }
   }
   
   if comment != "" {
       return fmt.Sprintf("// %s\n\nmodule %s\n\ngo %s\n", comment, moduleName, goVersion)
   }
   return fmt.Sprintf("module %s\n\ngo %s\n", moduleName, goVersion)
}

// generateGoWork creates a go.work file for a multi-module workspace.
func (g *DefaultContentGenerator) generateGoWork(relPath, comment string) string {
   // Using Go 1.24 as the default version
   goVersion := "1.24"
   
   // Try to get the actual Go version from the environment
   output, err := exec.Command("go", "version").Output()
   if err == nil {
       // Parse version from output like "go version go1.24.2 darwin/arm64"
       versionStr := string(output)
       versionParts := strings.Fields(versionStr)
       if len(versionParts) >= 3 {
           // Extract version number without "go" prefix
           versionFull := strings.TrimPrefix(versionParts[2], "go")
           // Take only major.minor (1.24 from 1.24.2)
           if dotIdx := strings.LastIndex(versionFull, "."); dotIdx > 0 {
               goVersion = versionFull[:dotIdx]
           } else {
               goVersion = versionFull
           }
       }
   }
   
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

// The legacy functions to maintain compatibility with existing code
var generators = map[string]FileGenerator{}

// RegisterGenerator associates an extension (e.g. ".go") with its generator.
func RegisterGenerator(ext string, gen FileGenerator) {
	generators[ext] = gen
}

// inferPkg derives the Go package name from relPath.
// Files under cmd/ or at the project root get package main;
// otherwise use the name of the parent directory.
func inferPkg(relPath string) string {
   dirPath := filepath.Dir(relPath)
   fileName := filepath.Base(relPath)
   
   // main.go files should always be package main
   if fileName == "main.go" {
       return "main"
   }
   
   // top-level files (Dir == ".") or cmd/* are main packages
   if strings.HasPrefix(relPath, "cmd/") || dirPath == "." {
       return "main"
   }
   
   return filepath.Base(dirPath)
}

// inferModuleName derives a Go module name from the relative path of a go.mod file.
// This is a best-effort guess based on common conventions.
func inferModuleName(relPath string) string {
   // Extract the directory where go.mod is located
   dir := filepath.Dir(relPath)
   
   // If it's in the root, use the current directory name
   if dir == "." {
       // Try to get the current git remote URL to determine a good module name
       output, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
       if err == nil {
           remoteURL := strings.TrimSpace(string(output))
           
           // Extract module name from common git URLs
           if strings.Contains(remoteURL, "github.com") {
               // Format: https://github.com/username/repo.git or git@github.com:username/repo.git
               urlParts := strings.Split(remoteURL, "/")
               if len(urlParts) >= 2 {
                   repoName := urlParts[len(urlParts)-1]
                   userName := urlParts[len(urlParts)-2]
                   
                   // Clean up username and repo name
                   repoName = strings.TrimSuffix(repoName, ".git")
                   if strings.Contains(userName, ":") {
                       userName = strings.Split(userName, ":")[1]
                   }
                   
                   return fmt.Sprintf("github.com/%s/%s", userName, repoName)
               }
           }
       }
       
       // Fallback: use current directory name
       cwd, err := exec.Command("pwd").Output()
       if err == nil {
           cwdStr := strings.TrimSpace(string(cwd))
           return filepath.Base(cwdStr)
       }
       
       return "example.com/mymodule"
   }
   
   // For nested modules, use the directory structure
   // This is a simple implementation and might need to be customized
   return "example.com/" + dir
}

// These functions are deprecated but kept for backward compatibility
func generateGo(relPath, comment string) string {
	gen := NewDefaultContentGenerator()
	return gen.generateGo(relPath, comment)
}

func generateGoWithRootPackage(relPath, comment, rootDirName string) string {
	name := filepath.Base(relPath)
   
	// Clean the rootDirName to be a valid Go package name
	// Remove path separators, spaces, and other invalid characters
	cleanPkg := strings.ToLower(rootDirName)
   
	// Replace invalid characters with underscores
	cleanPkg = strings.ReplaceAll(cleanPkg, "-", "_")
	cleanPkg = strings.ReplaceAll(cleanPkg, ".", "_")
   
	// Handle test_ prefix which is common in test directories
	if strings.HasPrefix(cleanPkg, "test_") {
		cleanPkg = strings.TrimPrefix(cleanPkg, "test_")
	}
   
	// If the package name becomes empty after cleaning, use a default
	if cleanPkg == "" {
		cleanPkg = "main"
	}
   
	if comment != "" {
		return fmt.Sprintf("// %s\n\npackage %s\n\nfunc main() {\n    // TODO: implement %s\n}\n", comment, cleanPkg, name)
	}
	return fmt.Sprintf("package %s\n\nfunc main() {\n    // TODO: implement %s\n}\n", cleanPkg, name)
}

// Legacy initialization for backward compatibility
func init() {
	// Register generators for different file types
	RegisterGenerator(".go", generateGo)
	RegisterGenerator("go.mod", NewDefaultContentGenerator().generateGoMod)
	RegisterGenerator("go.work", NewDefaultContentGenerator().generateGoWork)
	RegisterGenerator("go.sum", NewDefaultContentGenerator().generateGoSum)
}
