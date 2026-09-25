package main

import (
	"fmt"
	"os"

	"github.com/coccinella-labs/go-kit/internal/commands"
	"github.com/coccinella-labs/go-kit/internal/exitcode"
)

var Version = "1.0.0"

func main() {
	err := commands.Execute(Version)
	if err == nil {
		return
	}

	code := exitcode.ErrGeneral
	if exitcode.Is(err) {
		code = exitcode.Code(err)
	}

	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(code)
}
