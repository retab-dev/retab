package main

import (
	"fmt"
	"os"

	"github.com/retab-dev/retab/cli/cmd"
)

func main() {
	if err := cmd.ExecuteArgs(os.Args[1:]); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}
}
