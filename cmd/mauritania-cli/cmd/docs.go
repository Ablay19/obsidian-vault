package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/docs"
	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for the project",
	Long:  `Generate comprehensive documentation including README, API docs, and guides`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.Default()
		root := "."

		// Initialize analyzers and generators
		da := docs.NewDocAnalyzer(logger)
		rg := docs.NewReadmeGenerator()
		ag := docs.NewAPIGenerator()
		sg := docs.NewSetupGenerator()

		// Scan documents
		documents, err := da.ScanDocuments(root)
		if err != nil {
			logger.Printf("Failed to scan documents: %v", err)
			return
		}

		// Generate README
		err = rg.GenerateReadme(root)
		if err != nil {
			logger.Printf("Failed to generate README: %v", err)
			return
		}

		// Generate API docs
		err = ag.GenerateAPIDocs(root)
		if err != nil {
			logger.Printf("Failed to generate API docs: %v", err)
			return
		}

		// Generate setup guide
		err = sg.GenerateSetupGuide(root)
		if err != nil {
			logger.Printf("Failed to generate setup guide: %v", err)
			return
		}

		// Save documents
		dt := validation.NewDocumentTracker("cmd/mauritania-cli/internal/validation/data")
		err = dt.SaveDocuments(documents)
		if err != nil {
			logger.Printf("Failed to save documents: %v", err)
			return
		}

		fmt.Printf("Documentation generation complete. Generated %d documents.\n", len(documents))
	},
}
