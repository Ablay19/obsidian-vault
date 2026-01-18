package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/doppler"
)

// newTestCmd creates the test command
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests with Doppler integration",
		Long:  `Run end-to-end tests with Doppler environment variable injection`,
		Run: func(cmd *cobra.Command, args []string) {
			project, _ := cmd.Flags().GetString("project")
			config, _ := cmd.Flags().GetString("config")
			dopplerOnly, _ := cmd.Flags().GetBool("doppler-only")

			if err := runTestsWithDoppler(project, config, dopplerOnly, args); err != nil {
				log.Fatalf("Test execution failed: %v", err)
			}
		},
	}

	cmd.Flags().String("project", "bot", "Doppler project name")
	cmd.Flags().String("config", "dev", "Doppler config name")
	cmd.Flags().Bool("doppler-only", false, "Only run Doppler-related tests")

	return cmd
}

func runTestsWithDoppler(project, config string, dopplerOnly bool, args []string) error {
	manager := doppler.NewManager(project, config)

	// Set up fallbacks for testing
	manager.WithFallbacks(map[string]string{
		"TEST_DATABASE_URL":  "sqlite://:memory:",
		"TEST_REDIS_ADDR":    "localhost:6379",
		"TELEGRAM_BOT_TOKEN": "test_token",
		"WHATSAPP_API_KEY":   "test_key",
	})

	// Check if Doppler is available
	if manager.IsAvailable(context.Background()) {
		fmt.Println("✓ Doppler is available, loading secrets...")
		if err := manager.SetEnvironment(context.Background()); err != nil {
			fmt.Printf("⚠ Doppler setup failed, using fallbacks: %v\n", err)
		}
	} else {
		fmt.Println("⚠ Doppler not available, using fallback environment")
	}

	// Set test environment variables
	os.Setenv("GO_ENV", "test")
	os.Setenv("TEST_WITH_DOPPLER", "true")

	// Build test command
	testArgs := []string{"test"}

	// Add any additional test arguments
	testArgs = append(testArgs, args...)

	// Execute tests
	if dopplerOnly {
		// Run Doppler tests
		fmt.Println("Running Doppler tests...")
		dopplerArgs := []string{"test"}
		dopplerArgs = append(dopplerArgs, args...) // Only append user args, not testArgs
		dopplerArgs = append(dopplerArgs, "./internal/doppler/...")
		cmd1 := exec.Command("go", dopplerArgs...)
		cmd1.Stdout = os.Stdout
		cmd1.Stderr = os.Stderr
		if err := cmd1.Run(); err != nil {
			return fmt.Errorf("Doppler tests failed: %w", err)
		}

		// Run E2E tests
		fmt.Println("Running E2E tests...")
		e2eArgs := []string{"test"}
		e2eArgs = append(e2eArgs, args...) // Only append user args, not testArgs
		e2eArgs = append(e2eArgs, "./tests/e2e/...")
		cmd2 := exec.Command("go", e2eArgs...)
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err := cmd2.Run(); err != nil {
			return fmt.Errorf("E2E tests failed: %w", err)
		}

		fmt.Println("✅ All Doppler tests passed!")
		return nil
	} else {
		// Run all tests
		testArgs = append(testArgs, "./...")
		fmt.Printf("Running: go %s\n", strings.Join(testArgs, " "))

		cmd := exec.Command("go", testArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}
}
