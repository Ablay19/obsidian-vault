package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/ui"
)

// newResultCmd creates the result command
func newResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result [command-id]",
		Short: "Get command execution result",
		Long:  `Retrieve the result and output of a completed command execution.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commandID := args[0]

			resultContent := fmt.Sprintf("Command ID: %s\nExit Code: %s\nStatus: %s\n\nOutput:\n%s",
				ui.FormatID(commandID),
				ui.FormatStatus("0"),
				ui.FormatStatus("Success"),
				"Command executed successfully")

			ui.Println(ui.SuccessBox("Command Result", resultContent))

			// TODO: Implement actual result retrieval logic
			return nil
		},
	}

	return cmd
}
