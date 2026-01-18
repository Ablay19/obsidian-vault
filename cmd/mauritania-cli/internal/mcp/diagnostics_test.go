package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDiagnosticsTool(t *testing.T) {
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "diagnostics",
		},
	}

	result, err := handleDiagnosticsTool(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)

	content := result.Content[0]
	textContent, ok := content.(mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Diagnostics")
}
