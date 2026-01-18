package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleStatusTool(t *testing.T) {
	ctx := context.Background()

	// Test with empty arguments
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "status",
		},
	}

	result, err := handleStatusTool(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)

	// Check that result contains status information
	content := result.Content[0]
	textContent, ok := content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Status")
}
