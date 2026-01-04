package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

// PdfToText converts a PDF file to text using pdftotext command.
func PdfToText(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return string(output), nil
}

// ConvertMarkdownToPDF converts a Markdown file to PDF using pandoc and tectonic.
// This function requires pandoc and tectonic to be installed and in the system's PATH.
func ConvertMarkdownToPDF(markdownContent, outputPath string) error {
	// Step 1: Create a temporary Markdown file
	tmpFile, err := os.CreateTemp("", "note-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary markdown file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(markdownContent)); err != nil {
		return fmt.Errorf("failed to write to temporary markdown file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary markdown file: %w", err)
	}

	// Step 2: Run pandoc to convert Markdown to PDF
	// Using --pdf-engine=tectonic requires pandoc to be configured to find tectonic.
	// Make sure tectonic is in the PATH.
	cmd := exec.Command("pandoc", tmpFile.Name(), "-o", outputPath, "--pdf-engine=tectonic")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc conversion failed: %w\nOutput: %s", err, string(output))
	}

	zap.S().Info("Successfully converted Markdown to PDF using pandoc", "output_path", outputPath)
	return nil
}

// ensureDir ensures a directory exists, creating it if necessary.
func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// GetPDFPath generates a standardized path for a PDF file.
func GetPDFPath(noteTitle string) (string, error) {
	dir := "pdfs" // Or some other configurable directory
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s.pdf", noteTitle)), nil
}
