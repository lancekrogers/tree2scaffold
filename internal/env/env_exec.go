//go:build !wasip1

package env

import (
	"os"
	"os/exec"
	"strings"
)

// execEnv is the native Environment: it shells out to the host toolchain, VCS,
// and clipboard. It is excluded from the wasip1 build so os/exec never enters
// the WASI binary.
type execEnv struct{}

// New returns the native, exec-backed Environment.
func New() Environment { return execEnv{} }

// GoVersion parses `go version` output into a "major.minor" string.
func (execEnv) GoVersion() (string, error) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "", err
	}
	if v := parseGoMinor(string(out)); v != "" {
		return v, nil
	}
	return "", ErrUnsupported
}

// GitRemoteOriginURL returns the trimmed `git config --get remote.origin.url`.
func (execEnv) GitRemoteOriginURL() (string, error) {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Getwd uses os.Getwd directly (no `pwd` subprocess) so it works everywhere.
func (execEnv) Getwd() (string, error) { return os.Getwd() }

// Clipboard reads the macOS clipboard via pbpaste.
func (execEnv) Clipboard() ([]byte, error) { return exec.Command("pbpaste").Output() }

// parseGoMinor turns a `go version` line into a "major.minor" string, e.g.
// "go version go1.24.2 darwin/arm64" -> "1.24" and "go version go1.24 ..." ->
// "1.24". It returns "" when the version string cannot be parsed.
func parseGoMinor(raw string) string {
	fields := strings.Fields(raw)
	if len(fields) < 3 {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(fields[2], "go"), ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return parts[0]
}
