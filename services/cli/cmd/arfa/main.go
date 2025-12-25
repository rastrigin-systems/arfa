package main

import (
	"fmt"
	"os"

	"github.com/rastrigin-systems/arfa/services/cli/internal/commands"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
)

var version = "v0.2.0-dev"

func main() {
	// Create dependency injection container
	c := container.New()
	defer func() { _ = c.Close() }()

	if err := commands.NewRootCommand(version, c).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
