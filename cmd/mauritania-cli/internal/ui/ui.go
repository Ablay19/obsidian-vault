package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Title      lipgloss.Style
	Border     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
	Command    lipgloss.Style
	Output     lipgloss.Style
	Prompt     lipgloss.Style
}

func NewStyles() *Styles {
	s := &Styles{}

	// Base colors inspired by nushell
	primary := lipgloss.Color("#7EB2DD")   // Light blue
	secondary := lipgloss.Color("#61AFEF") // Blue
	success := lipgloss.Color("#98C379")   // Green
	error := lipgloss.Color("#E06C75")     // Red
	warning := lipgloss.Color("#E5C07B")   // Yellow
	info := lipgloss.Color("#56B6C2")      // Cyan
	border := lipgloss.Color("#3E4451")    // Border color

	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(primary).
		MarginBottom(1)

	s.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1)

	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(primary).
		Background(lipgloss.Color("#2C313A"))

	s.Unselected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ABB2BF"))

	s.Success = lipgloss.NewStyle().
		Foreground(success).
		Bold(true)

	s.Error = lipgloss.NewStyle().
		Foreground(error).
		Bold(true)

	s.Warning = lipgloss.NewStyle().
		Foreground(warning).
		Bold(true)

	s.Info = lipgloss.NewStyle().
		Foreground(info)

	s.Command = lipgloss.NewStyle().
		Foreground(primary).
		Bold(true)

	s.Output = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ABB2BF"))

	s.Prompt = lipgloss.NewStyle().
		Foreground(secondary).
		Bold(true)

	return s
}

// LogLevel represents different log levels
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarning
	LogError
	LogCommand
)

// LogMessage represents a log message with styling
type LogMessage struct {
	Level   LogLevel
	Message string
	Time    string
}

// FormatLog formats a log message with appropriate styling
func (s *Styles) FormatLog(msg LogMessage) string {
	var prefix string
	var style lipgloss.Style

	switch msg.Level {
	case LogSuccess:
		prefix = "✅"
		style = s.Success
	case LogError:
		prefix = "❌"
		style = s.Error
	case LogWarning:
		prefix = "⚠️"
		style = s.Warning
	case LogCommand:
		prefix = "🔧"
		style = s.Command
	default:
		prefix = "ℹ️"
		style = s.Info
	}

	timestamp := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")).Render(msg.Time)
	prefixStyled := style.Render(prefix)

	return fmt.Sprintf("%s %s %s", timestamp, prefixStyled, style.Render(msg.Message))
}

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Command     string
}

// FilterValue implements list.Item interface
func (i MenuItem) FilterValue() string { return i.Title }

// MainMenuModel represents the main menu
type MainMenuModel struct {
	list     list.Model
	styles   *Styles
	selected string
}

// NewMainMenu creates a new main menu
func NewMainMenu() MainMenuModel {
	styles := NewStyles()

	items := []list.Item{
		MenuItem{Title: "Send Command", Description: "Send a command via social media", Command: "send"},
		MenuItem{Title: "View Status", Description: "Check transport status", Command: "status"},
		MenuItem{Title: "Configure", Description: "Configure application settings", Command: "config"},
		MenuItem{Title: "Monitor", Description: "Monitor commands and transports", Command: "monitor"},
		MenuItem{Title: "Security", Description: "Security and authentication settings", Command: "security"},
		MenuItem{Title: "Help", Description: "Show help and usage information", Command: "help"},
		MenuItem{Title: "Exit", Description: "Exit the application", Command: "exit"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Mauritania CLI - Remote Development Interface"
	l.Styles.Title = styles.Title
	l.Styles.FilterPrompt = styles.Prompt
	l.Styles.FilterCursor = styles.Command

	return MainMenuModel{
		list:   l,
		styles: styles,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(MenuItem); ok {
				m.selected = i.Command
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainMenuModel) View() string {
	return m.styles.Border.Render(m.list.View())
}

// GetSelected returns the selected command
func (m MainMenuModel) GetSelected() string {
	return m.selected
}

// CommandInputModel represents command input interface
type CommandInputModel struct {
	textInput textinput.Model
	spinner   spinner.Model
	styles    *Styles
	senderID  string
	command   string
	executing bool
	result    string
}

func NewCommandInput(senderID string) CommandInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter command to execute..."
	ti.CharLimit = 500
	ti.Width = 80
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF"))

	return CommandInputModel{
		textInput: ti,
		spinner:   s,
		styles:    NewStyles(),
		senderID:  senderID,
		executing: false,
	}
}

func (m CommandInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CommandInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.executing {
				m.command = m.textInput.Value()
				m.executing = true
				m.result = ""
				return m, m.spinner.Tick
			}
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		}
	}

	if m.executing {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m CommandInputModel) View() string {
	var view string

	if m.senderID != "" {
		view += m.styles.Info.Render(fmt.Sprintf("📱 Sender: %s\n\n", m.senderID))
	}

	if m.executing {
		view += m.styles.Command.Render("Executing command...\n")
		view += m.spinner.View() + "\n\n"
		if m.result != "" {
			view += m.styles.Output.Render("Result:\n" + m.result)
		}
	} else {
		view += m.styles.Prompt.Render("Command: ") + "\n"
		view += m.textInput.View() + "\n\n"
		view += m.styles.Info.Render("💡 Press Enter to execute, Esc to go back, Ctrl+C to quit")
	}

	return m.styles.Border.Render(view)
}

// GetCommand returns the entered command
func (m CommandInputModel) GetCommand() string {
	return m.command
}

// SetResult sets the execution result
func (m CommandInputModel) SetResult(result string) CommandInputModel {
	m.result = result
	m.executing = false
	return m
}

// StatusModel represents status display
type StatusModel struct {
	styles     *Styles
	statusInfo string
}

func NewStatusModel(statusInfo string) StatusModel {
	return StatusModel{
		styles:     NewStyles(),
		statusInfo: statusInfo,
	}
}

func (m StatusModel) Init() tea.Cmd {
	return nil
}

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m StatusModel) View() string {
	title := m.styles.Title.Render("📊 System Status")
	content := m.styles.Output.Render(m.statusInfo)
	footer := m.styles.Info.Render("\nPress 'q', 'esc', or 'ctrl+c' to go back")

	return m.styles.Border.Render(title + "\n\n" + content + footer)
}

// ConfigModel represents configuration display
type ConfigModel struct {
	styles      *Styles
	configInfo  string
	currentView string
	textInput   textinput.Model
}

func NewConfigModel(configInfo string) ConfigModel {
	ti := textinput.New()
	ti.Placeholder = "Enter config key to edit (or 'back' to return)..."
	ti.CharLimit = 100
	ti.Width = 60

	return ConfigModel{
		styles:      NewStyles(),
		configInfo:  configInfo,
		currentView: "view",
		textInput:   ti,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e":
			if m.currentView == "view" {
				m.currentView = "edit"
				m.textInput.Focus()
				return m, textinput.Blink
			}
		case "esc":
			if m.currentView == "edit" {
				m.currentView = "view"
				m.textInput.Blur()
				m.textInput.SetValue("")
			} else {
				return m, tea.Quit
			}
		case "enter":
			if m.currentView == "edit" {
				input := m.textInput.Value()
				if input == "back" {
					m.currentView = "view"
					m.textInput.Blur()
					m.textInput.SetValue("")
				}
				// Here you would handle config editing
			}
		}
	}

	if m.currentView == "edit" {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m ConfigModel) View() string {
	var view string

	if m.currentView == "edit" {
		view += m.styles.Title.Render("⚙️  Configuration Editor")
		view += "\n\n"
		view += m.styles.Prompt.Render("Config key: ") + "\n"
		view += m.textInput.View() + "\n\n"
		view += m.styles.Info.Render("💡 Enter config key to edit, or 'back' to return to view mode")
	} else {
		view += m.styles.Title.Render("⚙️  Current Configuration")
		view += "\n\n"
		view += m.styles.Output.Render(m.configInfo)
		view += "\n\n"
		view += m.styles.Info.Render("Press 'e' to edit, 'q' or 'esc' to go back")
	}

	return m.styles.Border.Render(view)
}

// HelpModel represents help display
type HelpModel struct {
	styles *Styles
}

func NewHelpModel() HelpModel {
	return HelpModel{
		styles: NewStyles(),
	}
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m HelpModel) View() string {
	title := m.styles.Title.Render("🆘 Mauritania CLI Help")

	content := m.styles.Output.Render(`
Mauritania CLI enables remote development in low-connectivity regions through social media.

NAVIGATION:
• Use arrow keys to navigate menus
• Press Enter to select options
• Press 'q', 'esc', or 'ctrl+c' to go back

COMMANDS:
• Send Command: Execute commands remotely via social media
• View Status: Check transport and system status
• Configure: Manage application settings
• Monitor: View command history and metrics
• Security: Configure authentication and permissions

TRANSPORTS:
• WhatsApp: Primary transport with webhook support
• Telegram: Bot-based messaging
• Facebook: Messenger integration
• SM APOS Shipper: Secure command execution

For more detailed help, see: mauritania-cli help
`)

	footer := m.styles.Info.Render("\nPress 'q', 'esc', or 'ctrl+c' to go back")

	return m.styles.Border.Render(title + content + footer)
}
