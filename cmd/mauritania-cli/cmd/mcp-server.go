package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"obsidian-automation/internal/mcp"
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

func init() {
	rootCmd.AddCommand(mcpServerCmd)

	mcpServerCmd.Flags().String("transport", "stdio", "Transport mechanism (stdio, http)")
	mcpServerCmd.Flags().String("host", "localhost", "Host for HTTP transport")
	mcpServerCmd.Flags().String("port", "8080", "Port for HTTP transport")
}
