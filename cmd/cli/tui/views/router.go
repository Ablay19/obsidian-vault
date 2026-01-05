package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ViewType represents different view types
type ViewType int

const (
	MenuView ViewType = iota
	StatusView
	AIProvidersView
	UsersView
)

// Route represents a route in CLI
type Route struct {
	Name     string
	ViewType ViewType
	Model    tea.Model
	MenuText string
	Icon     string
}

// Router manages navigation between different views
type Router struct {
	routes    map[ViewType]*Route
	current   ViewType
	styles    Styles
	backStack []ViewType
}

// NewRouter creates a new router with default routes
func NewRouter() *Router {
	styles := NewStyles(DefaultPalette())

	return &Router{
		routes:    make(map[ViewType]*Route),
		current:   MenuView,
		styles:    styles,
		backStack: make([]ViewType, 0),
	}
}

// RegisterRoute registers a new route
func (r *Router) RegisterRoute(viewType ViewType, route *Route) {
	r.routes[viewType] = route
}

// NavigateTo navigates to a specific view
func (r *Router) NavigateTo(viewType ViewType) tea.Cmd {
	if route, exists := r.routes[viewType]; exists {
		// Add current view to back stack if not the same
		if viewType != r.current {
			r.backStack = append(r.backStack, r.current)
		}
		r.current = viewType

		// Initialize the model using a type assertion for Init method
		type InitInterface interface {
			Init() tea.Cmd
		}

		if model, ok := route.Model.(InitInterface); ok {
			return model.Init()
		}
	}
	return nil
}

// GoBack navigates to the previous view
func (r *Router) GoBack() tea.Cmd {
	if len(r.backStack) > 0 {
		// Pop from back stack
		lastIndex := len(r.backStack) - 1
		previousView := r.backStack[lastIndex]
		r.backStack = r.backStack[:lastIndex]

		r.current = previousView
		return nil
	}
	return r.NavigateTo(MenuView) // Default to menu if no back stack
}

// GetCurrent returns the current route
func (r *Router) GetCurrent() *Route {
	if route, exists := r.routes[r.current]; exists {
		return route
	}
	return nil
}

// GetCurrentModel returns the current model
func (r *Router) GetCurrentModel() tea.Model {
	if route := r.GetCurrent(); route != nil {
		return route.Model
	}
	return nil
}

// Update updates the current model
func (r *Router) Update(msg tea.Msg) tea.Cmd {
	currentRoute := r.GetCurrent()
	if currentRoute == nil {
		return nil
	}

	// Handle navigation keys
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if r.current != MenuView {
				return r.GoBack()
			}
		case "ctrl+c":
			return tea.Quit
		}
	}

	// Update current model
	if model, ok := currentRoute.Model.(tea.Model); ok {
		var cmd tea.Cmd
		currentRoute.Model, cmd = model.Update(msg)
		return cmd
	}

	return nil
}

// View renders the current view with header and footer
func (r *Router) View() string {
	currentRoute := r.GetCurrent()
	if currentRoute == nil {
		return "No route selected"
	}

	// Build header
	header := r.styles.Header.Render("Obsidian Bot CLI")

	// Build breadcrumbs
	breadcrumbs := r.buildBreadcrumbs()

	// Build main content
	content := ""
	if model, ok := currentRoute.Model.(interface{ View() string }); ok {
		content = model.View()
	}

	// Build help text
	help := r.buildHelpText()

	// Combine all sections
	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s",
		header,
		breadcrumbs,
		content,
		help,
	)
}

// buildBreadcrumbs creates navigation breadcrumbs
func (r *Router) buildBreadcrumbs() string {
	if r.current == MenuView {
		return ""
	}

	breadcrumbs := []string{"Home"}

	// Add current view name
	if currentRoute := r.GetCurrent(); currentRoute != nil {
		breadcrumbs = append(breadcrumbs, currentRoute.Name)
	}

	result := ""
	for i, crumb := range breadcrumbs {
		if i > 0 {
			result += " → "
		}
		result += r.styles.MenuInactive.Render(crumb)
	}

	return result
}

// buildHelpText creates contextual help text
func (r *Router) buildHelpText() string {
	helpItems := []string{"Ctrl+C: Quit"}

	if r.current != MenuView {
		helpItems = append(helpItems, "Esc/Q: Back to Menu")
	}

	switch r.current {
	case StatusView:
		helpItems = append(helpItems, "R: Refresh Status")
	case UsersView:
		helpItems = append(helpItems, "R: Refresh Users")
	case AIProvidersView:
		helpItems = append(helpItems, "R: Refresh Providers")
	case MenuView:
		helpItems = append(helpItems, "↑/↓: Navigate | Enter: Select")
	}

	help := ""
	for i, item := range helpItems {
		if i > 0 {
			help += " • "
		}
		help += r.styles.Help.Render(item)
	}

	return help
}

// GetRoutes returns all registered routes
func (r *Router) GetRoutes() map[ViewType]*Route {
	return r.routes
}
