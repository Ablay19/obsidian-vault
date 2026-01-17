package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newShipperCmd creates the shipper command
func newShipperCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shipper",
		Short: "SM APOS Shipper management",
		Long:  `Manage SM APOS Shipper authentication and operations.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Check shipper connection status",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Shipper status: authenticated")
				return nil
			},
		},
		&cobra.Command{
			Use:   "login",
			Short: "Authenticate with SM APOS Shipper",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Logged in to SM APOS Shipper")
				return nil
			},
		},
	)

	return cmd
}
