package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/russross/blackfriday/v2"
)

// convertMarkdownToHTML converts a Markdown string to an HTML string.
func convertMarkdownToHTML(markdownContent string) string {
	html := blackfriday.Run([]byte(markdownContent))
	return string(html)
}

// convertHTMLToPDF converts an HTML string to a PDF byte slice using headless Chrome.
func convertHTMLToPDF(htmlContent string) ([]byte, error) {
	// Create a new context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Create a temporary file for the HTML content
	tmpfile, err := ioutil.TempFile("", "htmltopdf.*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(htmlContent)); err != nil {
		tmpfile.Close()
		return nil, fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	var pdfBuffer []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("file://%s", tmpfile.Name())),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().Do(ctx)
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp failed: %w", err)
	}

	log.Println("Successfully converted HTML to PDF")
	return pdfBuffer, nil
}
