package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/coverage"
	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate functionality and generate coverage reports",
	Long:  `Run comprehensive validation including test execution and coverage analysis`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.Default()

		dataDir := "cmd/mauritania-cli/internal/validation/data"
		root := "."

		// Initialize trackers
		ft := validation.NewFeatureTracker(dataDir)
		dt := validation.NewDocumentTracker(dataDir)
		dirT := validation.NewDirectoryTracker(dataDir)

		// Scan features
		scanner := validation.NewFeatureScanner(root)
		features, err := scanner.ScanFeatures()
		if err != nil {
			log.Fatal(err)
		}

		// Execute tests
		executor := validation.NewTestExecutor(logger)
		ca := coverage.NewCoverageAnalyzer(logger)
		for i := range features {
			err = executor.ExecuteTests(&features[i])
			if err != nil {
				logger.Printf("Test execution failed: %v", err)
			}
			err = ca.UpdateFeatureCoverage(&features[i])
			if err != nil {
				logger.Printf("Coverage update failed: %v", err)
			}
		}

		// Save features
		err = ft.SaveFeatures(features)
		if err != nil {
			log.Fatal(err)
		}

		// Generate report
		reporter := validation.NewValidationReporter()
		reportPath := filepath.Join(dataDir, "validation_report.txt")
		err = reporter.GenerateReport(features, reportPath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Validation complete. Report saved to %s\n", reportPath)

		// Load and display
		loaded, _ := ft.LoadFeatures()
		for _, f := range loaded {
			fmt.Printf("Feature: %s - Status: %s - Coverage: %.2f\n", f.Name, f.Status, f.TestCoverage)
		}

		// Placeholder for docs and dirs
		docs, _ := dt.LoadDocuments()
		fmt.Printf("Documents: %d\n", len(docs))
		dirs, _ := dirT.LoadDirectories()
		fmt.Printf("Directories: %d\n", len(dirs))
	},
}
