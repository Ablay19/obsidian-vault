package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleErrorMetricsTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		window   string
		expected string
	}{
		{
			name:     "default window",
			window:   "day",
			expected: "Error metrics",
		},
		{
			name:     "hour window",
			window:   "hour",
			expected: "Error metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "error_metrics",
					Arguments: map[string]interface{}{
						"window": tt.window,
					},
				},
			}

			result, err := handleErrorMetricsTool(ctx, req)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Content)

			content := result.Content[0]
			textContent, ok := content.(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, textContent.Text, tt.expected)
		})
	}
}
