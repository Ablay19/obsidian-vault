package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleTestConnectionTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		transport string
		expected  string
	}{
		{
			name:      "whatsapp transport",
			transport: "whatsapp",
			expected:  "Connection test",
		},
		{
			name:      "telegram transport",
			transport: "telegram",
			expected:  "Connection test",
		},
		{
			name:      "facebook transport",
			transport: "facebook",
			expected:  "Connection test",
		},
		{
			name:      "shipper transport",
			transport: "shipper",
			expected:  "Connection test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "test_connection",
					Arguments: map[string]interface{}{
						"transport": tt.transport,
					},
				},
			}

			result, err := handleTestConnectionTool(ctx, req)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Content)

			content := result.Content[0]
			textContent, ok := content.(mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, textContent.Text, tt.expected)
		})
	}
}
