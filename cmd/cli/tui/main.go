package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the CLI application
func Run() {
	p := tea.NewProgram(NewApp())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
