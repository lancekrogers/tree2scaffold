package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
	"github.com/lancekrogers/tree2scaffold/pkg/scaffold"
)

func askConfirm() bool {
	fmt.Print("Proceed? [y/N]: ")
	var resp string
	if _, err := fmt.Scanln(&resp); err != nil {
		return false
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes"
}

func main() {
	root := flag.String("root", ".", "project root")
	dryRun := flag.Bool("dry-run", false, "show what would be created and ask")
	alwaysYes := flag.Bool("yes", false, "skip confirmation prompt")
	flag.Parse()

	// read from stdin or fallback to clipboard
	var input io.Reader = os.Stdin
	fi, _ := os.Stdin.Stat()
	if fi.Mode()&os.ModeCharDevice != 0 {
		// no pipe, use pbpaste
		out, err := exec.Command("pbpaste").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read clipboard:", err)
			os.Exit(1)
		}
		input = bytes.NewReader(out)
	}

	nodes, err := parser.Parse(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	// preview
	if *dryRun {
		fmt.Println("☑️  Will create:")
		for _, n := range nodes {
			if n.IsDir {
				fmt.Printf("    dir:  %s\n", n.Path)
			} else {
				fmt.Printf("    file: %s\n", n.Path)
			}
		}
		if !*alwaysYes && !askConfirm() {
			fmt.Println("Aborted.")
			return
		}
	}

	// apply with progress callback
	err = scaffold.Apply(*root, nodes, func(path string, isDir bool) {
		if isDir {
			fmt.Printf("📁 mkdir %s\n", path)
		} else {
			fmt.Printf("📝 write %s\n", path)
		}
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "scaffold error:", err)
		os.Exit(1)
	}
}
