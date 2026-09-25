package main

import (
	"fmt"
	"os"

	"github.com/coccinella-labs/go-kit/internal/commands"
	"github.com/spf13/cobra/doc"
)

func main() {
	root := commands.NewRootCmd("dev")
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, "./docs"); err != nil {
		fmt.Fprintf(os.Stderr, "error generating docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated docs in ./docs/")
}
