package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// scaffoldBin is the tree2scaffold executable under test. It is built once for
// the whole package so the suite exercises the working tree rather than
// whichever copy happens to be installed on the developer's PATH.
var scaffoldBin string

func TestMain(m *testing.M) {
	code, err := runSuite(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runSuite(m *testing.M) (int, error) {
	dir, err := os.MkdirTemp("", "tree2scaffold-bin")
	if err != nil {
		return 0, fmt.Errorf("create temp bin dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	scaffoldBin = filepath.Join(dir, "tree2scaffold")
	if runtime.GOOS == "windows" {
		scaffoldBin += ".exe"
	}

	build := exec.Command("go", "build", "-o", scaffoldBin, "../cmd/tree2scaffold")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return 0, fmt.Errorf("build tree2scaffold: %w", err)
	}

	return m.Run(), nil
}
