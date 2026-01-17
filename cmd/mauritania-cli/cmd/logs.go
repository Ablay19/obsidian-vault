package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newLogsCmd creates the logs command
func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Demonstrate logging capabilities",
		Long:  `Show various logging features and formatted output similar to Nushell.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonMode, _ := cmd.Flags().GetBool("json")
			logger := utils.NewLogger("DEMO")
			logger.SetJSON(jsonMode)

			// Banner
			logger.Banner("Nushell-Style Logging Demo")

			// Section headers
			logger.Section("Log Levels")

			// Different log levels with clean formatting
			logger.Info("This is an info message with structured data", "user", "alice", "action", "login")
			logger.Success("Operation completed successfully", "items_processed", 42)
			logger.Warn("Warning: disk space running low", "available_gb", 5.2)
			logger.Error("Failed to connect to database", fmt.Errorf("connection timeout"))
			logger.Debug("Debug information", "trace_id", "abc-123", "correlation_id", "xyz-789")

			logger.Section("Structured Data")

			// Structured logging examples
			logger.Info("Service health check",
				"service", "ai-worker",
				"status", "healthy",
				"uptime", "2h 15m",
				"cpu_percent", 45.2,
				"memory_mb", 128)

			logger.Info("Network request",
				"method", "POST",
				"url", "/api/v1/render",
				"status_code", 200,
				"response_time_ms", 150,
				"bytes_sent", 2048)

			logger.Section("Progress Indicators")

			// Progress simulation with different scenarios
			logger.Info("Starting batch processing", "total_items", 5)

			scenarios := []struct {
				name  string
				total int
			}{
				{"Fast operations", 3},
				{"Slow operations", 4},
				{"Mixed operations", 5},
			}

			for _, scenario := range scenarios {
				logger.Info(fmt.Sprintf("Processing %s", scenario.name), "count", scenario.total)
				for i := 1; i <= scenario.total; i++ {
					logger.Progress(i, scenario.total, scenario.name)
				}
			}

			logger.Section("Table Display")

			// Enhanced table display
			headers := []string{"Service", "Status", "CPU %", "Memory", "Uptime", "Health"}
			rows := [][]string{
				{"ai-worker", "running", "45.2", "128MB", "2h 15m", "healthy"},
				{"manim-renderer", "running", "12.8", "256MB", "1h 30m", "healthy"},
				{"api-proxy", "stopped", "0.0", "0MB", "0s", "unknown"},
				{"scheduler", "running", "5.1", "64MB", "3h 45m", "healthy"},
				{"monitor", "running", "2.3", "32MB", "4h 20m", "healthy"},
			}
			logger.Table(headers, rows)

			logger.Section("Command Output")

			// Command output simulation with more realistic examples
			logger.CommandOutput("kubectl get pods", true,
				"NAME                          READY   STATUS    RESTARTS   AGE\nai-worker-7f8b9c4d5-x7y2z     1/1     Running   0          2h\nmanim-renderer-6d7e8f9g0-a1b2c 1/1     Running   0          1h\napi-proxy-5c6d7e8f9g-h3i4j     0/1     Pending   0          5m", 1200000000)

			logger.CommandOutput("docker build .", false,
				"Sending build context to Docker daemon  2.048kB\nStep 1/5 : FROM golang:1.21-alpine\n ---> c9c8e7d8b3c2\nStep 2/5 : WORKDIR /app\n ---> Using cache\n ---> a1b2c3d4e5f6\nStep 3/5 : COPY go.mod .\n ---> Using cache\n ---> f6e5d4c3b2a1\nStep 4/5 : RUN go mod download\n ---> Running in abc123def456\nERROR: go.mod:3: invalid module path", 8500000000)

			logger.CommandOutput("npm test", true,
				"\n> mauritania-cli@1.0.0 test /app\n> vitest run\n\n ✓ utils/logger.test.ts (2)\n ✓ utils/network.test.ts (5)\n ✓ cmd/send.test.ts (3)\n\nTest Files  3 passed (3)\n     Tests  10 passed (10)\n      Time  2.45s", 2450000000)

			return nil
		},
	}

	return cmd
}
