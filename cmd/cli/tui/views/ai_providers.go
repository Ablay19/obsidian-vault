package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

)

// AIProvidersModel represents the AI providers view
type AIProvidersModel struct {
	table      table.Model
	styles     Styles
	loading    bool
	lastUpdate time.Time
	providers  []AIProvider
}

// AIProvider represents an AI provider in the table
type AIProvider struct {
	Name         string
	Model        string
	Status       string
	KeyID        string
	Enabled      bool
	LastUsed     time.Time
	ResponseTime int64 // milliseconds
}

// NewAIProviders creates a new AI providers model
func NewAIProviders(styles Styles) AIProvidersModel {
	columns := []table.Column{
		{Title: "Provider", Width: 15},
		{Title: "Model", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Key ID", Width: 15},
		{Title: "Response", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	return AIProvidersModel{
		table:      t,
		styles:     styles,
		loading:    true,
		lastUpdate: time.Now(),
		providers:  []AIProvider{},
	}
}

// Init initializes the AI providers model
func (m AIProvidersModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		m.fetchProvidersCmd(),
	)
}

// Update updates the AI providers model
func (m AIProvidersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, nil // Return to main menu
		case "r", "R":
			m.loading = true
			return m, m.fetchProvidersCmd()
		}

	case providersMsg:
		m.providers = msg.providers
		m.loading = false
		m.lastUpdate = time.Now()
		m.updateTableRows()
		return m, nil

	case errorMsgMsg:
		// Error handling is done in View method
		return m, nil
	}

	// Update table
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the AI providers view
func (m AIProvidersModel) View() string {
	content := m.styles.Card.Render(
		m.styles.Header.Render("🤖 AI Providers") + "\n\n" +
			m.renderContent(),
	)

	footer := m.styles.Footer.Render(
		fmt.Sprintf("Last updated: %s | R: Refresh Providers | Q: Back to Menu",
			m.lastUpdate.Format("15:04:05"),
		),
	)

	return content + "\n" + footer
}

// renderContent renders the main content
func (m AIProvidersModel) renderContent() string {
	if m.loading {
		return m.styles.Loading.Render("🔄 Loading AI providers...")
	}

	if len(m.providers) == 0 {
		return m.styles.Muted.Render("No AI providers configured. Please check your environment variables.")
	}

	// Render table
	m.updateTableRows()
	return m.table.View()
}

// updateTableRows updates the table with provider data
func (m AIProvidersModel) updateTableRows() {
	rows := make([]table.Row, len(m.providers))
	for i, provider := range m.providers {
		status := "🔴 Disabled"
		statusStyle := m.styles.Error

		if provider.Enabled {
			status = "🟢 Active"
			statusStyle = m.styles.Success
		}

		responseTime := "N/A"
		if provider.ResponseTime > 0 {
			responseTime = fmt.Sprintf("%dms", provider.ResponseTime)
			if provider.ResponseTime < 100 {
				responseTime = m.styles.Success.Render(responseTime)
			} else if provider.ResponseTime < 500 {
				responseTime = m.styles.Warning.Render(responseTime)
			} else {
				responseTime = m.styles.Error.Render(responseTime)
			}
		}

		rows[i] = table.Row{
			provider.Name,
			provider.Model,
			statusStyle.Render(status),
			m.styles.Muted.Render(provider.KeyID),
			responseTime,
		}
	}

	m.table.SetRows(rows)
}

// fetchProvidersCmd creates a command to fetch AI providers
func (m AIProvidersModel) fetchProvidersCmd() tea.Cmd {
	return func() tea.Msg {
		providers, err := m.fetchRealProviders()
		if err != nil {
			return errorMsgMsg{err: err.Error()}
		}
		return providersMsg{providers: providers}
	}
}

// fetchRealProviders fetches real AI provider data
// In a real implementation, this would call the actual AI service
func (m AIProvidersModel) fetchRealProviders() ([]AIProvider, error) {
	// Mock data for demonstration
	// TODO: Replace with actual AI service calls
	now := time.Now()
	mockProviders := []AIProvider{
		{
			Name:         "Gemini",
			Model:        "gemini-1.5-flash",
			Status:       "active",
			KeyID:        "gem-key-1",
			Enabled:      true,
			LastUsed:     now.Add(-time.Minute * 30),
			ResponseTime: 85,
		},
		{
			Name:         "Groq",
			Model:        "llama-3.1-8b",
			Status:       "active",
			KeyID:        "groq-key-1",
			Enabled:      true,
			LastUsed:     now.Add(-time.Minute * 15),
			ResponseTime: 120,
		},
		{
			Name:         "OpenRouter",
			Model:        "gpt-3.5-turbo",
			Status:       "inactive",
			KeyID:        "openrouter-key-1",
			Enabled:      false,
			LastUsed:     now.Add(-time.Hour * 24),
			ResponseTime: 0,
		},
		{
			Name:         "HuggingFace",
			Model:        "mixtral-8x7b",
			Status:       "active",
			KeyID:        "hf-key-1",
			Enabled:      true,
			LastUsed:     now.Add(-time.Minute * 5),
			ResponseTime: 250,
		},
	}

	return mockProviders, nil
}

// Command messages
type providersMsg struct {
	providers []AIProvider
}


