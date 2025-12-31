package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// convertMarkdownToPDF converts a Markdown string to a PDF byte slice using pandoc and tectonic.
func convertMarkdownToPDF(markdownContent string) ([]byte, error) {
	// Create a temporary file for the Markdown content
	tmpMarkdownFile, err := ioutil.TempFile("", "markdowntopdf.*.md")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp markdown file: %w", err)
	}
	defer os.Remove(tmpMarkdownFile.Name())

	if _, err := tmpMarkdownFile.Write([]byte(markdownContent)); err != nil {
		tmpMarkdownFile.Close()
		return nil, fmt.Errorf("failed to write to temp markdown file: %w", err)
	}
	if err := tmpMarkdownFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp markdown file: %w", err)
	}

	// Create a temporary file for the PDF output
	tmpPDFFile, err := ioutil.TempFile("", "output.*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp pdf file: %w", err)
	}
	tmpPDFFile.Close() // Close it immediately, pandoc will write to it.
	defer os.Remove(tmpPDFFile.Name())

	cmd := exec.Command("pandoc", tmpMarkdownFile.Name(), "-o", tmpPDFFile.Name(), "--pdf-engine=tectonic")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pandoc failed: %w\n%s", err, string(output))
	}

	pdfBuffer, err := ioutil.ReadFile(tmpPDFFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp pdf file: %w", err)
	}

	log.Println("Successfully converted Markdown to PDF using pandoc")
	return pdfBuffer, nil
}