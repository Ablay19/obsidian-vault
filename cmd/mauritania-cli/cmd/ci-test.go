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

// newCITestCmd creates the ci-test command
func newCITestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci-test",
		Short: "Run CI/CD tests with Doppler integration",
		Long:  `Run automated tests in CI/CD environment with Doppler service tokens`,
		Run: func(cmd *cobra.Command, args []string) {
			project, _ := cmd.Flags().GetString("project")
			config, _ := cmd.Flags().GetString("config")
			coverage, _ := cmd.Flags().GetBool("coverage")

			if err := runCITests(project, config, coverage); err != nil {
				log.Fatalf("CI test execution failed: %v", err)
			}
		},
	}

	cmd.Flags().String("project", "bot", "Doppler project name")
	cmd.Flags().String("config", "e2e", "Doppler config name")
	cmd.Flags().Bool("coverage", true, "Generate coverage report")

	return cmd
}

func runCITests(project, config string, coverage bool) error {
	fmt.Printf("🚀 Starting CI/CD tests with Doppler integration\n")
	fmt.Printf("Project: %s, Config: %s\n", project, config)

	// Initialize Doppler manager
	manager := doppler.NewManager(project, config)

	// Verify Doppler availability
	if !manager.IsAvailable(context.Background()) {
		return fmt.Errorf("Doppler is not available - check DOPPLER_TOKEN and network")
	}

	fmt.Println("✓ Doppler connection verified")

	// Load and set environment
	if err := manager.SetEnvironment(context.Background()); err != nil {
		fmt.Printf("⚠ Doppler environment setup failed, using fallbacks: %v\n", err)
	}

	// Set CI environment variables
	os.Setenv("CI", "true")
	os.Setenv("GO_ENV", "test")
	os.Setenv("TEST_WITH_DOPPLER", "true")

	// Run tests
	testArgs := []string{"test", "-v", "-race"}

	if coverage {
		testArgs = append(testArgs, "-coverprofile=coverage.out")
	}

	// Focus on E2E and Doppler tests in CI
	testArgs = append(testArgs, "./internal/doppler/...", "./tests/e2e/...")

	fmt.Printf("Running: go %s\n", strings.Join(testArgs, " "))

	cmd := exec.Command("go", testArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("✅ All tests passed!")

	// Generate coverage report if requested
	if coverage {
		if err := generateCoverageReport(); err != nil {
			fmt.Printf("⚠ Coverage report generation failed: %v\n", err)
		}
	}

	return nil
}

func generateCoverageReport() error {
	// Convert coverage profile to HTML
	cmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("📊 Coverage report generated: coverage.html")
	return nil
}
