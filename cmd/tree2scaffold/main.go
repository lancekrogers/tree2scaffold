// Package main provides the command-line interface for tree2scaffold, a tool
// that converts ASCII tree diagrams into file system directories and files.
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

// Command-line options
type options struct {
	root           string
	dryRun         bool
	alwaysYes      bool
	debug          bool
	forceOverwrite bool
}

// askConfirm prompts the user for confirmation and returns their response
func askConfirm() bool {
	fmt.Print("Proceed? [y/N]: ")
	var resp string
	if _, err := fmt.Scanln(&resp); err != nil {
		return false
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes"
}

// getInput returns an io.Reader with the input to process
// It either reads from stdin or falls back to clipboard
func getInput() (io.Reader, error) {
	// Check if stdin has data
	fi, _ := os.Stdin.Stat()
	if fi.Mode()&os.ModeCharDevice == 0 {
		// Data is being piped in
		return os.Stdin, nil
	}
	
	// No pipe, try to use pbpaste
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read clipboard: %w", err)
	}
	return bytes.NewReader(out), nil
}

// preprocessInput applies any necessary preprocessing to the input
func preprocessInput(input io.Reader, debug bool) (io.Reader, error) {
	if !debug {
		return input, nil
	}
	
	// For debug mode, print the raw input
	inputBytes, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	
	fmt.Println("=== Input ===")
	fmt.Println(string(inputBytes))
	fmt.Println("=== End Input ===")
	
	return bytes.NewReader(inputBytes), nil
}

// previewNodes prints a preview of what will be created
func previewNodes(nodes []parser.Node) {
	fmt.Println("☑️  Will create:")
	for _, n := range nodes {
		if n.IsDir {
			fmt.Printf("    dir:  %s\n", n.Path)
		} else {
			fmt.Printf("    file: %s\n", n.Path)
		}
	}
}

// debugNodes prints detailed node information in debug mode
func debugNodes(nodes []parser.Node) {
	fmt.Println("=== Parsed Nodes ===")
	for i, n := range nodes {
		fmt.Printf("%d: Path=%s, IsDir=%v, Comment=%s\n", i, n.Path, n.IsDir, n.Comment)
	}
	fmt.Println("=== End Parsed Nodes ===")
}

// parseFlags parses command-line flags into an options structure
func parseFlags() options {
	opts := options{}
	
	// Define standard flags
	flag.StringVar(&opts.root, "root", ".", "project root directory")
	flag.BoolVar(&opts.dryRun, "dry-run", false, "show what would be created and ask")
	flag.BoolVar(&opts.alwaysYes, "yes", false, "skip confirmation prompt")
	flag.BoolVar(&opts.debug, "debug", false, "output debug information")
	flag.BoolVar(&opts.forceOverwrite, "force", false, "force overwrite of existing files that conflict with directories")
	
	// Add a special shortcut flag for dry-run (abbreviated 'd')
	dShortcut := flag.Bool("d", false, "shortcut for --dry-run")
	
	// Parse flags
	flag.Parse()
	
	// Apply the shortcut if used
	if *dShortcut {
		opts.dryRun = true
	}
	
	return opts
}

// run executes the main program logic
func run(opts options) error {
	// Get the input
	input, err := getInput()
	if err != nil {
		return err
	}
	
	// Preprocess the input if needed
	input, err = preprocessInput(input, opts.debug)
	if err != nil {
		return err
	}
	
	// Parse the input into nodes
	nodes, err := parser.Parse(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	
	// Debug mode - print the parsed nodes
	if opts.debug {
		debugNodes(nodes)
	}
	
	// Preview what will be created
	previewNodes(nodes)
	
	// Create a scaffolder
	var s scaffold.Scaffolder
	if opts.forceOverwrite {
		s = scaffold.NewScaffolderWithForce()
	} else {
		s = scaffold.NewScaffolder()
	}
	
	// Pre-validate, especially for hidden files
	if !opts.forceOverwrite {
		if err := s.Validate(opts.root, nodes); err != nil {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Options:\n")
			fmt.Fprintf(os.Stderr, "  1. Remove conflicting files manually before running again\n")
			fmt.Fprintf(os.Stderr, "  2. Use the -force flag to overwrite conflicting files\n")
			return err
		}
	} else if opts.debug {
		fmt.Println("Note: Force mode enabled - will attempt to overwrite conflicting files")
	}
	
	// Handle dry run mode
	if opts.dryRun {
		if !opts.alwaysYes && !askConfirm() {
			fmt.Println("Aborted.")
			return nil
		}
	}
	
	// Apply the scaffold and report progress
	err = s.Apply(opts.root, nodes, func(path string, isDir bool) {
		if isDir {
			fmt.Printf("📁 mkdir %s\n", path)
		} else {
			fmt.Printf("📝 write %s\n", path)
		}
	})
	
	if err != nil {
		return fmt.Errorf("scaffold error: %w", err)
	}
	
	return nil
}

// main is the entry point for the application
func main() {
	// Parse command-line flags
	opts := parseFlags()
	
	// Run the application
	err := run(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
