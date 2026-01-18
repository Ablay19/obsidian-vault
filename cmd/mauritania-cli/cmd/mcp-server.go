package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/mcp"
)

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for AI diagnostics",
	Long:  `Start the Model Context Protocol server to expose diagnostic tools for AI assistants`,
	Run: func(cmd *cobra.Command, args []string) {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")

		if err := mcp.StartServer(transport, host, port); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	},
}

// newMCPServerCmd creates the MCP server command
func newMCPServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp-server",
		Short: "Start MCP server for AI diagnostics",
		Long:  `Start the Model Context Protocol server to expose diagnostic tools for AI assistants`,
		Run: func(cmd *cobra.Command, args []string) {
			transport, _ := cmd.Flags().GetString("transport")
			port, _ := cmd.Flags().GetString("port")
			host, _ := cmd.Flags().GetString("host")

			if err := mcp.StartServer(transport, host, port); err != nil {
				log.Fatalf("MCP server error: %v", err)
			}
		},
	}

	cmd.Flags().String("transport", "stdio", "Transport mechanism (stdio, http)")
	cmd.Flags().String("host", "localhost", "Host for HTTP transport")
	cmd.Flags().String("port", "8080", "Port for HTTP transport")

	return cmd
}
