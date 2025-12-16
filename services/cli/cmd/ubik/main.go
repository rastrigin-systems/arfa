package main

import (
	"fmt"
	"os"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands"
)

var version = "v0.2.0-dev"

func main() {
	if err := commands.NewRootCommand(version).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
