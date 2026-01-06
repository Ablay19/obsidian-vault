package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusModel represents the status view
type StatusModel struct {
	styles     Styles
	spinner    spinner.Model
	loading    bool
	status     *StatusData
	lastUpdate time.Time
	error      error
}

// StatusData contains system status information
type StatusData struct {
	BotStatus      string           `json:"bot_status"`
	AIService      string           `json:"ai_service"`
	DatabaseStatus string           `json:"database_status"`
	LastActivity   time.Time        `json:"last_activity"`
	Services       []*ServiceStatus `json:"services"`
	Metrics        *StatusMetrics   `json:"metrics"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// StatusMetrics contains system metrics
type StatusMetrics struct {
	Uptime       string `json:"uptime"`
	MessageCount int    `json:"message_count"`
	ErrorCount   int    `json:"error_count"`
	MemoryUsage  string `json:"memory_usage"`
}

// NewStatus creates a new status model
func NewStatus(styles Styles) StatusModel {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.Loading

	return StatusModel{
		styles:     styles,
		spinner:    s,
		loading:    true,
		lastUpdate: time.Now(),
	}
}

// Init initializes the status model
func (m StatusModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchStatusCmd(),
	)
}

// Update updates the status model
func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			m.loading = true
			m.error = nil
			return m, tea.Batch(
				m.spinner.Tick,
				m.fetchStatusCmd(),
			)
		case "q", "esc", "ctrl+c":
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statusDataMsg:
		m.status = msg.data
		m.loading = false
		m.error = nil
		m.lastUpdate = time.Now()
		return m, nil

	case statusErrorMsg:
		m.error = msg.err
		m.loading = false
		m.lastUpdate = time.Now()
		return m, nil
	}

	return m, nil
}

// View renders the status view
func (m StatusModel) View() string {
	content := m.styles.Card.Render(
		m.styles.Header.Render("📊 System Status") + "\n\n" +
			m.renderContent(),
	)

	footer := m.styles.Footer.Render(
		fmt.Sprintf("Last updated: %s | R: Refresh | Q: Back",
			m.lastUpdate.Format("15:04:05"),
		),
	)

	return content + "\n" + footer
}

// renderContent renders the main status content
func (m StatusModel) renderContent() string {
	if m.loading {
		return m.styles.Loading.Render(
			"🔄 Fetching system status..." + "\n" +
				m.spinner.View(),
		)
	}

	if m.error != nil {
		return m.styles.Error.Render(
			"❌ Error fetching status:" + "\n" +
				m.error.Error(),
		)
	}

	if m.status == nil {
		return m.styles.Muted.Render("No status data available")
	}

	// Build status cards
	cards := []string{
		m.renderStatusCard("🤖", "Bot Core", m.status.BotStatus),
		m.renderStatusCard("🧠", "AI Service", m.status.AIService),
		m.renderStatusCard("🗄", "Database", m.status.DatabaseStatus),
		m.renderActivityCard(),
	}

	result := ""
	for i, card := range cards {
		if i > 0 {
			result += "\n"
		}
		result += card
	}

	return result
}

// renderStatusCard renders a single status card
func (m StatusModel) renderStatusCard(icon, title, status string) string {
	statusStyle := m.styles.Muted
	statusIcon := "❌"

	switch status {
	case "Healthy", "Running", "Active":
		statusStyle = m.styles.Success
		statusIcon = "✅"
	case "Warning", "Degraded":
		statusStyle = m.styles.Warning
		statusIcon = "⚠️"
	case "Starting", "Loading":
		statusStyle = m.styles.Info
		statusIcon = "🔄"
	}

	return m.styles.Card.Render(
		icon + " " + m.styles.Subtitle.Render(title) + "\n" +
			statusStyle.Render(statusIcon+" "+status) + "\n" +
			m.renderServices(),
	)
}

// renderActivityCard renders the activity information
func (m StatusModel) renderActivityCard() string {
	activity := "Never"
	if !m.status.LastActivity.IsZero() {
		activity = fmt.Sprintf("%s (%s ago)",
			m.status.LastActivity.Format("2006-01-02 15:04:05"),
			time.Since(m.status.LastActivity).Round(time.Minute).String(),
		)
	}

	return m.styles.Card.Render(
		"📈 " + m.styles.Subtitle.Render("Last Activity") + "\n" +
			m.styles.Value.Render(activity) + "\n" +
			m.renderMetrics(),
	)
}

// renderServices renders additional services
func (m StatusModel) renderServices() string {
	if len(m.status.Services) == 0 {
		return ""
	}

	result := ""
	for i, service := range m.status.Services {
		if i > 0 {
			result += "\n"
		}

		status := "●"
		switch service.Status {
		case "Healthy", "Running":
			status = "●"
		case "Warning":
			status = "◐"
		case "Error":
			status = "✖"
		}

		result += fmt.Sprintf("  %s %s: %s",
			status,
			m.styles.Label.Render(service.Name),
			m.styles.Muted.Render(service.Status),
		)
	}

	return result
}

// renderMetrics renders system metrics
func (m StatusModel) renderMetrics() string {
	if m.status.Metrics == nil {
		return ""
	}

	metrics := []string{
		fmt.Sprintf("Uptime: %s", m.status.Metrics.Uptime),
		fmt.Sprintf("Messages: %d", m.status.Metrics.MessageCount),
		fmt.Sprintf("Errors: %d", m.status.Metrics.ErrorCount),
	}

	if m.status.Metrics.MemoryUsage != "" {
		metrics = append(metrics, fmt.Sprintf("Memory: %s", m.status.Metrics.MemoryUsage))
	}

	result := ""
	for i, metric := range metrics {
		if i > 0 {
			result += " • "
		}
		result += m.styles.Muted.Render(metric)
	}

	return result
}

// fetchStatusCmd fetches real status data
func (m StatusModel) fetchStatusCmd() tea.Cmd {
	return func() tea.Msg {
		// Mock data for demonstration
		// TODO: Replace with actual status fetching logic
		time.Sleep(2 * time.Second) // Simulate network delay

		mockStatus := &StatusData{
			BotStatus:      "Running",
			AIService:      "Healthy",
			DatabaseStatus: "Connected",
			LastActivity:   time.Now().Add(-5 * time.Minute),
			Services: []*ServiceStatus{
				{Name: "Telegram API", Status: "Healthy", LastCheck: time.Now()},
				{Name: "WhatsApp API", Status: "Healthy", LastCheck: time.Now()},
				{Name: "AI Provider 1", Status: "Healthy", LastCheck: time.Now()},
			},
			Metrics: &StatusMetrics{
				Uptime:       "2d 10h",
				MessageCount: 1234,
				ErrorCount:   5,
				MemoryUsage:  "128MB",
			},
		}

		// Simulate potential error
		// if time.Now().Second()%2 == 0 {
		// 	return statusErrorMsg{err: fmt.Errorf("simulated fetch error")}
		// }

		return statusDataMsg{data: mockStatus}
	}
}

// Command messages
type statusDataMsg struct {
	data *StatusData
}

type statusErrorMsg struct {
	err error
}
