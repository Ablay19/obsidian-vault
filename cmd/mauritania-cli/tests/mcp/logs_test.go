package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLogsTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		level    string
		hours    float64
		expected string
	}{
		{
			name:     "default parameters",
			level:    "error",
			hours:    24,
			expected: "Logs retrieval not implemented yet",
		},
		{
			name:     "custom level",
			level:    "info",
			hours:    1,
			expected: "Logs retrieval not implemented yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "logs",
					Arguments: map[string]interface{}{
						"level": tt.level,
						"hours": tt.hours,
					},
				},
			}

			result, err := handleLogsTool(ctx, req)

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
