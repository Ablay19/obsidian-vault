package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"obsidian-automation/cmd/cli/tui/views"
)

// AppModel represents the main application state
type AppModel struct {
	router      *views.Router
	styles      views.Styles
	quitting    bool
	initialized bool
	width       int
	height      int
}

// Message represents application-level messages
type Message struct {
	Type    string
	Content interface{}
}

// NewApp creates a new application model
func NewApp() *AppModel {
	// Create styles and router
	styles := views.NewStyles(views.DefaultPalette())
	router := views.NewRouter()

	app := &AppModel{
		router:      router,
		styles:      styles,
		quitting:    false,
		initialized: false,
	}

	// Register all routes with their models
	app.registerRoutes()

	return app
}

// registerRoutes sets up all application routes
func (app *AppModel) registerRoutes() {
	// Create route models with shared styles
	menuModel := views.NewMenu(app.styles)
	statusModel := views.NewStatus(app.styles)
	usersModel := views.NewUsers(app.styles)
	aiProvidersModel := views.NewAIProviders(app.styles)

	// Register routes with router
	app.router.RegisterRoute(views.MenuView, &views.Route{
		Name:     "Main Menu",
		ViewType: views.MenuView,
		Model:    menuModel,
		MenuText: "Main Menu",
		Icon:     "🏠",
	})

	app.router.RegisterRoute(views.StatusView, &views.Route{
		Name:     "System Status",
		ViewType: views.StatusView,
		Model:    statusModel,
		MenuText: "Bot Status",
		Icon:     "📊",
	})

	app.router.RegisterRoute(views.UsersView, &views.Route{
		Name:     "User Management",
		ViewType: views.UsersView,
		Model:    usersModel,
		MenuText: "User Management",
		Icon:     "👥",
	})

	app.router.RegisterRoute(views.AIProvidersView, &views.Route{
		Name:     "AI Providers",
		ViewType: views.AIProvidersView,
		Model:    aiProvidersModel,
		MenuText: "AI Providers",
		Icon:     "🤖",
	})
}

// Init initializes the application
func (app *AppModel) Init() tea.Cmd {
	app.initialized = true
	return tea.Batch(
		tea.WindowSize(),
		app.router.NavigateTo(views.MenuView),
	)
}

// Update handles application updates
func (app *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		app.width = sizeMsg.Width
		app.height = sizeMsg.Height
		return app, nil
	}

	// Handle application-level messages
	if message, ok := msg.(Message); ok {
		return app, app.handleAppMessage(message)
	}

	// Handle navigation from menu
	if navMsg, ok := msg.(views.NavigationMsg); ok {
		cmd := app.router.NavigateTo(navMsg.Target)
		return app, cmd
	}

	// Update current view through router
	cmd := app.router.Update(msg)

	// Update router state back to app model
	// Note: We would need to expose router state updates here

	return app, cmd
}

// View renders the application
func (app *AppModel) View() string {
	if app.quitting {
		return app.styles.Footer.Render("Goodbye! 👋")
	}

	if !app.initialized {
		return app.styles.Loading.Render("🔄 Initializing...")
	}

	// Render through router for consistent header/footer
	return app.router.View()
}

// handleAppMessage handles application-specific messages
func (app *AppModel) handleAppMessage(msg Message) tea.Cmd {
	switch msg.Type {
	case "quit":
		app.quitting = true
		return tea.Quit

	case "navigate":
		if viewType, ok := msg.Content.(views.ViewType); ok {
			return app.router.NavigateTo(viewType)
		}
	case "refresh_current":
		// Send refresh to current view
		return app.sendToCurrentView("refresh", nil)

	case "error":
		if errorMsg, ok := msg.Content.(string); ok {
			return app.sendToCurrentView("error", errorMsg)
		}

	case "success":
		if successMsg, ok := msg.Content.(string); ok {
			return app.sendToCurrentView("success", successMsg)
		}
	}

	return nil
}

// sendToCurrentView sends a message to the currently active view
func (app *AppModel) sendToCurrentView(msgType string, content interface{}) tea.Cmd {
	currentRoute := app.router.GetCurrent()
	if currentRoute == nil {
		return nil
	}

	// This would require adding message handling to view interfaces
	// For now, this is a placeholder for the pattern
	return func() tea.Msg {
		return Message{
			Type:    msgType,
			Content: content,
		}
	}
}

// ShowError displays an error message
func (app *AppModel) ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return Message{
			Type:    "error",
			Content: err.Error(),
		}
	}
}

// ShowSuccess displays a success message
func (app *AppModel) ShowSuccess(message string) tea.Cmd {
	return func() tea.Msg {
		return Message{
			Type:    "success",
			Content: message,
		}
	}
}

// NavigateTo navigates to a specific view
func (app *AppModel) NavigateTo(viewType views.ViewType) tea.Cmd {
	return func() tea.Msg {
		return Message{
			Type:    "navigate",
			Content: viewType,
		}
	}
}
