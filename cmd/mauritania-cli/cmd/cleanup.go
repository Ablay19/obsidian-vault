package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/cleanup"
	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

var dryRun bool

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up unnecessary files and directories",
	Long:  `Analyze and remove temporary files, build artifacts, and other unnecessary files`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.Default()
		root := "."

		// Initialize
		ct := cleanup.NewCleanupTargets(logger)
		ce := cleanup.NewCleanupExecutor(logger)
		ca := cleanup.NewCleanupAnalyzer(logger)

		// Load gitignore
		gitignorePatterns, err := ct.LoadGitignore(root)
		if err != nil {
			logger.Printf("Failed to load .gitignore: %v", err)
			return
		}

		// Identify targets
		targets, err := ct.IdentifyTargets(root, gitignorePatterns)
		if err != nil {
			logger.Printf("Failed to identify targets: %v", err)
			return
		}

		// Execute cleanup
		err = ce.ExecuteCleanup(targets, dryRun)
		if err != nil {
			logger.Printf("Cleanup failed: %v", err)
			return
		}

		// Update directories
		directories, err := ca.AnalyzeDirectories(root)
		if err != nil {
			logger.Printf("Failed to analyze directories: %v", err)
			return
		}

		dt := validation.NewDirectoryTracker("cmd/mauritania-cli/internal/validation/data")
		err = dt.SaveDirectories(directories)
		if err != nil {
			logger.Printf("Failed to save directories: %v", err)
			return
		}

		if dryRun {
			fmt.Printf("Dry run complete. Would remove %d files.\n", len(targets))
		} else {
			fmt.Printf("Cleanup complete. Removed %d files.\n", len(targets))
		}
	},
}

func init() {
	cleanupCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed without actually removing")
}
