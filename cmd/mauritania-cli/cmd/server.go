package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newServerCmd creates the server command
func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server",
		Long:  `Start the embedded HTTP server for API access and webhooks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			fmt.Printf("Starting server on port %d...\n", port)
			fmt.Println("Server started successfully")

			// TODO: Implement actual server startup logic
			return nil
		},
	}

	cmd.Flags().IntP("port", "p", 3001, "Port to run the server on")

	return cmd
}
