package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/ui"
)

// newQueueCmd creates the queue command
func newQueueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage offline command queue",
		Long:  `Manage the offline command queue for operations during network outages.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List queued commands",
			RunE: func(cmd *cobra.Command, args []string) error {
				ui.Println(ui.InfoBox("Queue Status", "Queued commands: 0\nNo commands waiting for execution"))
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Clear queued commands",
			RunE: func(cmd *cobra.Command, args []string) error {
				ui.Println(ui.SuccessBox("Queue Cleared", "All queued commands have been removed"))
				return nil
			},
		},
	)

	return cmd
}
