package main

import (
	"cmdforge/cli/cmd"
	"cmdforge/core"
	"fmt"
	"os"
)

func main() {
	// Initialize core registries (blocks, items, effects, etc.)
	if err := core.LoadData(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading Minecraft registry data: %v\n", err)
		os.Exit(1)
	}

	// Execute Cobra CLI
	cmd.Execute()
}
