package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorPalette(t *testing.T) {
	palette := DefaultPalette()

	// Test that all colors are defined
	assert.NotNil(t, palette.Primary)
	assert.NotNil(t, palette.Secondary)
	assert.NotNil(t, palette.Success)
	assert.NotNil(t, palette.Warning)
	assert.NotNil(t, palette.Error)
	assert.NotNil(t, palette.Info)
	assert.NotNil(t, palette.Muted)
	assert.NotNil(t, palette.Background)
	assert.NotNil(t, palette.Surface)
	assert.NotNil(t, palette.Border)
	assert.NotNil(t, palette.Text)
	assert.NotNil(t, palette.TextSecondary)
}
