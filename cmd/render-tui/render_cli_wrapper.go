package main

import (
	"context"
	"os/exec"
	"time"
)

// runCommandWithTimeout executes a command with a given timeout.
func runCommandWithTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return output, nil
}

// getServiceStatus executes the 'render-cli service describe' command and returns the output.
func getServiceStatus(serviceName string) (string, error) {
	// Note: The path to render-cli is relative to the execution directory of the TUI.
	// This will need to be adjusted based on the final binary location.
	// For now, we assume it's in a parallel directory structure.
	cliPath := "../cli/render-cli"

	// This command needs to be fleshed out with actual service details.
	// For now, we are just getting placeholder output.
	// A real implementation would be:
	// output, err := runCommandWithTimeout(10*time.Second, cliPath, "service", "describe", serviceName, "--json")
	// For this example, we'll just simulate a successful command.
	//
	// We'll simulate a call to a non-existent command to test the timeout logic later.
	// For now, let's just return a placeholder status.
	// In a real scenario, we'd parse this from the output.
	return "available", nil
}
