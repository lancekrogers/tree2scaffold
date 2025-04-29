package scaffold

import (
   "fmt"
   "os/exec"
   "path/filepath"
   "strings"
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
	// Register generators for different file types
	RegisterGenerator(".go", generateGo)
	RegisterGenerator("go.mod", generateGoMod)
	RegisterGenerator("go.work", generateGoWork)
	RegisterGenerator("go.sum", generateGoSum)
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
	".mod":  {"// ", ""}, // go.mod files use Go-style comments
	".work": {"// ", ""}, // go.work files use Go-style comments
	".sum":  {"// ", ""}, // go.sum files use Go-style comments
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

// generateGoMod creates a go.mod file with the current Go version.
func generateGoMod(relPath, comment string) string {
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
func generateGoWork(relPath, comment string) string {
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
func generateGoSum(relPath, comment string) string {
   if comment != "" {
       return fmt.Sprintf("// %s\n// This file will be automatically populated when dependencies are added to go.mod\n", comment)
   }
   return "// This file will be automatically populated when dependencies are added to go.mod\n"
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
