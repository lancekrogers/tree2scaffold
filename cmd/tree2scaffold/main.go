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
	debug := flag.Bool("debug", false, "output debug information")
	forceOverwrite := flag.Bool("force", false, "force overwrite of existing files that conflict with directories")
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

	// Print raw input for debugging
	if *debug {
		inputBytes, _ := io.ReadAll(input)
		fmt.Println("=== Input ===")
		fmt.Println(string(inputBytes))
		fmt.Println("=== End Input ===")
		input = bytes.NewReader(inputBytes)
	}

	nodes, err := parser.Parse(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}
	
	if *debug {
		fmt.Println("=== Parsed Nodes ===")
		for i, n := range nodes {
			fmt.Printf("%d: Path=%s, IsDir=%v, Comment=%s\n", i, n.Path, n.IsDir, n.Comment)
		}
		fmt.Println("=== End Parsed Nodes ===")
	}

	// preview
	fmt.Println("☑️  Will create:")
	for _, n := range nodes {
		if n.IsDir {
			fmt.Printf("    dir:  %s\n", n.Path)
		} else {
			fmt.Printf("    file: %s\n", n.Path)
		}
	}
	// Set the global force mode
	scaffold.ForceMode = *forceOverwrite
	
	// Pre-validate, especially for hidden files
	if !*forceOverwrite {
		if err := scaffold.Validate(*root, nodes); err != nil {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Options:\n")
			fmt.Fprintf(os.Stderr, "  1. Remove conflicting files manually before running again\n")
			fmt.Fprintf(os.Stderr, "  2. Use the -force flag to overwrite conflicting files\n")
			os.Exit(1)
		}
	} else if *debug {
		fmt.Println("Note: Force mode enabled - will attempt to overwrite conflicting files")
	}
	
	if *dryRun {
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
