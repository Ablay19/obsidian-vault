package cmd

import (
	"obsidian-automation/cmd/mauritania-cli/internal/ui"

	"github.com/spf13/cobra"
)

// newTUICmd creates the TUI command
func newTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Start interactive terminal user interface",
		Long:  `Start the Mauritania CLI in interactive TUI mode with colored output and menus.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := ui.NewTUIApplication()
			return app.Start()
		},
	}

	return cmd
}
