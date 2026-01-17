package main

import (
	"fmt"
	"os"

	"obsidian-automation/cmd/mauritania-cli/cmd"
	"obsidian-automation/cmd/mauritania-cli/internal/ui"
)

func main() {
	// Check for TUI mode flag
	tuiMode := false
	args := os.Args[1:]

	// Check if --tui flag is present
	for i, arg := range args {
		if arg == "--tui" {
			tuiMode = true
			// Remove --tui from args
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if tuiMode {
		runTUI()
	} else {
		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs(args)

		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runTUI() {
	// Initialize TUI application
	app := ui.NewTUIApplication()
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
		os.Exit(1)
	}
}
