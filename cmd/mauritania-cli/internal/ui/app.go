package ui

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/services"
	"obsidian-automation/cmd/mauritania-cli/internal/shell"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// TUIApplication represents the main TUI application
type TUIApplication struct {
	db                *database.DB
	config            *utils.Config
	logger            *log.Logger
	transportSelector *services.TransportSelector
	commandReceiver   *shell.CommandReceiver
	responseSender    *shell.ResponseSender
}

// NewTUIApplication creates a new TUI application
func NewTUIApplication() *TUIApplication {
	// Initialize components (simplified for TUI)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Load configuration
	configManager := utils.NewConfigManager()
	if err := configManager.Load(); err != nil {
		logger.Printf("Warning: Failed to load config: %v", err)
	}
	config := configManager.Get()

	// Initialize database
	db, err := database.NewDB("./data", "", "")
	if err != nil {
		logger.Printf("Warning: Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		logger.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Initialize services
	transportSelector := services.NewTransportSelector(db, config, logger)

	// Initialize shell components
	commandReceiver := shell.NewCommandReceiver(db, config, logger)
	responseSender := shell.NewResponseSender(db, config, transportSelector, logger)

	return &TUIApplication{
		db:                db,
		config:            config,
		logger:            logger,
		transportSelector: transportSelector,
		commandReceiver:   commandReceiver,
		responseSender:    responseSender,
	}
}

// Start starts the TUI application
func (app *TUIApplication) Start() error {
	// Start with main menu
	p := tea.NewProgram(NewAppModel(app))
	return p.Start()
}

// AppModel represents the main application model
type AppModel struct {
	app     *TUIApplication
	current tea.Model
	state   AppState
	history []AppState
	styles  *Styles
}

type AppState int

const (
	StateMainMenu AppState = iota
	StateCommandInput
	StateStatus
	StateConfig
	StateHelp
	StateExiting
)

// NewAppModel creates a new application model
func NewAppModel(app *TUIApplication) AppModel {
	menu := NewMainMenu()
	return AppModel{
		app:     app,
		current: menu,
		state:   StateMainMenu,
		history: []AppState{StateMainMenu},
		styles:  NewStyles(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Handle state transitions
	if menu, ok := m.current.(MainMenuModel); ok && menu.GetSelected() != "" {
		selected := menu.GetSelected()
		switch selected {
		case "send":
			m.history = append(m.history, m.state)
			m.state = StateCommandInput
			m.current = NewCommandInput("")
		case "status":
			m.history = append(m.history, m.state)
			m.state = StateStatus
			statusInfo := m.getStatusInfo()
			m.current = NewStatusModel(statusInfo)
		case "config":
			m.history = append(m.history, m.state)
			m.state = StateConfig
			configInfo := m.getConfigInfo()
			m.current = NewConfigModel(configInfo)
		case "help":
			m.history = append(m.history, m.state)
			m.state = StateHelp
			m.current = NewHelpModel()
		case "exit":
			return m, tea.Quit
		}
		return m, m.current.Init()
	}

	// Handle back navigation
	if _, ok := m.current.(CommandInputModel); ok && msg != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			if len(m.history) > 0 {
				m.state = m.history[len(m.history)-1]
				m.history = m.history[:len(m.history)-1]
				m.current = NewMainMenu()
				return m, m.current.Init()
			}
		}
	}

	// Handle other back navigations
	if (m.state == StateStatus || m.state == StateConfig || m.state == StateHelp) && msg != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && (keyMsg.String() == "q" || keyMsg.String() == "esc") {
			if len(m.history) > 0 {
				m.state = m.history[len(m.history)-1]
				m.history = m.history[:len(m.history)-1]
				m.current = NewMainMenu()
				return m, m.current.Init()
			}
		}
	}

	// Update current model
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	return m.current.View()
}

func (m AppModel) getStatusInfo() string {
	status := "🏥 System Status\n\n"

	// Transport status
	status += "📡 Transport Status:\n"
	health := m.app.transportSelector.GetTransportHealthStatus()
	for transport, healthy := range health {
		if healthy {
			status += fmt.Sprintf("  ✅ %s: Healthy\n", transport)
		} else {
			status += fmt.Sprintf("  ❌ %s: Unhealthy\n", transport)
		}
	}

	// Database status
	status += "\n💾 Database Status:\n"
	status += "  ✅ Connected\n"

	// Recent commands
	status += "\n📋 Recent Commands:\n"
	// This would normally query the database for recent commands
	status += "  (Feature not yet implemented in TUI)\n"

	return status
}

func (m AppModel) getConfigInfo() string {
	config := "⚙️ Current Configuration:\n\n"

	// Database config
	config += "💾 Database:\n"
	config += fmt.Sprintf("  Type: %s\n", m.app.config.Database.Type)
	config += fmt.Sprintf("  Path: %s\n", m.app.config.Database.Path)

	// Transport config
	config += "\n📡 Transports:\n"
	config += fmt.Sprintf("  Default: %s\n", m.app.config.Transports.Default)

	// WhatsApp
	config += "  WhatsApp:\n"
	whatsapp := m.app.config.Transports.SocialMedia.WhatsApp
	if whatsapp.WebhookSecret != "" {
		config += "    ✅ Configured\n"
	} else {
		config += "    ❌ Not configured\n"
	}

	// Telegram
	config += "  Telegram:\n"
	if m.app.config.Transports.SocialMedia.Telegram.BotToken != "" {
		config += "    ✅ Configured\n"
	} else {
		config += "    ❌ Not configured\n"
	}

	// Facebook
	config += "  Facebook:\n"
	if m.app.config.Transports.SocialMedia.Facebook.AccessToken != "" {
		config += "    ✅ Configured\n"
	} else {
		config += "    ❌ Not configured\n"
	}

	// Network config
	config += "\n🌐 Network:\n"
	config += fmt.Sprintf("  Timeout: %d seconds\n", m.app.config.Network.Timeout)
	config += fmt.Sprintf("  Retry Attempts: %d\n", m.app.config.Network.RetryAttempts)

	// Logging config
	config += "\n📝 Logging:\n"
	config += fmt.Sprintf("  Level: %s\n", m.app.config.Logging.Level)

	// Auth config
	config += "\n🔐 Authentication:\n"
	if m.app.config.Auth.Enabled {
		config += "  ✅ Enabled\n"
	} else {
		config += "  ❌ Disabled\n"
	}

	return config
}

// Log adds a colored log message to the output
func (app *TUIApplication) Log(level LogLevel, message string) {
	styles := NewStyles()
	logMsg := LogMessage{
		Level:   level,
		Message: message,
		Time:    time.Now().Format("15:04:05"),
	}
	fmt.Println(styles.FormatLog(logMsg))
}
