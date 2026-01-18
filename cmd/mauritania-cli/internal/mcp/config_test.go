package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleConfigTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		transport string
		expected  string
	}{
		{
			name:      "no specific transport",
			transport: "",
			expected:  "Current sanitized configuration",
		},
		{
			name:      "specific transport",
			transport: "whatsapp",
			expected:  "Configuration for whatsapp transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			if tt.transport != "" {
				args["transport"] = tt.transport
			}

			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "config",
					Arguments: args,
				},
			}

			result, err := handleConfigTool(ctx, req)

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
