package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lancekrogers/tree2scaffold/pkg/parser"
	"github.com/lancekrogers/tree2scaffold/pkg/scaffold"
)

func main() {
	root := flag.String("root", ".", "project root")
	flag.Parse()

	nodes, err := parser.Parse(os.Stdin) // returns []Node{Path, IsDir, Comment}
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	if err := scaffold.Apply(*root, nodes); err != nil {
		fmt.Fprintln(os.Stderr, "scaffold error:", err)
		os.Exit(1)
	}
}
