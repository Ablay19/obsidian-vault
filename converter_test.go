package main

import (
	"testing"
)

func TestConvertMarkdownToHTML(t *testing.T) {
	testCases := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Simple H1",
			markdown: "# Hello World",
			expected: "<h1>Hello World</h1>\n",
		},
		{
			name:     "Simple Paragraph",
			markdown: "This is a paragraph.",
			expected: "<p>This is a paragraph.</p>\n",
		},
		{
			name:     "Bold Text",
			markdown: "**bold text**",
			expected: "<p><strong>bold text</strong></p>\n",
		},
		{
			name:     "Italic Text",
			markdown: "*italic text*",
			expected: "<p><em>italic text</em></p>\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := convertMarkdownToHTML(tc.markdown)
			if actual != tc.expected {
				t.Errorf("Expected %q, but got %q", tc.expected, actual)
			}
		})
	}
}
