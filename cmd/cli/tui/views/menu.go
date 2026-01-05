package views

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// MenuModel represents the main menu
type MenuModel struct {
	list     list.Model
	styles   Styles
	quitting bool
}

// MenuItem represents a menu option
type MenuItem struct {
	Name     string
	ViewType ViewType
	Icon     string
}

func (i MenuItem) FilterValue() string { return i.Name }

// MenuDelegate handles rendering menu items
type MenuDelegate struct {
	styles Styles
}

func NewMenuDelegate(styles Styles) MenuDelegate {
	return MenuDelegate{styles: styles}
}

func (d MenuDelegate) Height() int                               { return 1 }
func (d MenuDelegate) Spacing() int                              { return 0 }
func (d MenuDelegate) Update(msg tea.Msg, l *list.Model) tea.Cmd { return nil }
func (d MenuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	str := ""
	if item.Icon != "" {
		str = fmt.Sprintf("%s %s", item.Icon, item.Name)
	} else {
		str = item.Name
	}

	// Check if this item is selected
	if index == m.Index() {
		str = d.styles.MenuActive.Render("▶ " + str)
	} else {
		str = d.styles.MenuInactive.Render("  " + str)
	}

	io.WriteString(w, str)
}

// NewMenu creates a new menu model
func NewMenu(styles Styles) MenuModel {
	items := []list.Item{
		MenuItem{
			Name:     "Bot Status",
			ViewType: StatusView,
			Icon:     "📊",
		},
		MenuItem{
			Name:     "AI Providers",
			ViewType: AIProvidersView,
			Icon:     "🤖",
		},
		MenuItem{
			Name:     "User Management",
			ViewType: UsersView,
			Icon:     "👥",
		},
		MenuItem{
			Name:     "WhatsApp Integration",
			ViewType: ViewType(4), // WhatsApp view
			Icon:     "📱",
		},
		MenuItem{
			Name:     "Configuration",
			ViewType: ViewType(5), // Config view
			Icon:     "⚙️",
		},
	}

	const defaultWidth = 30

	l := list.New(items, NewMenuDelegate(styles), defaultWidth, 14)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.MenuTitle
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return MenuModel{
		list:     l,
		styles:   styles,
		quitting: false,
	}
}

// Init initializes the menu model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update updates the menu model based on messages
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item := m.list.SelectedItem(); item != nil {
				if menuItem, ok := item.(MenuItem); ok {
					// Return selection command for router
					return m, tea.Sequence(
						func() tea.Msg {
							return NavigationMsg{
								Target: menuItem.ViewType,
								Item:   menuItem,
							}
						},
					)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the menu
func (m MenuModel) View() string {
	if m.quitting {
		return m.styles.Footer.Render("Goodbye! 👋")
	}

	content := m.list.View()
	return content
}

// GetSelectedItem returns the selected menu item
func (m MenuModel) GetSelectedItem() *MenuItem {
	if item := m.list.SelectedItem(); item != nil {
		if menuItem, ok := item.(MenuItem); ok {
			return &menuItem
		}
	}
	return nil
}

// NavigationMsg represents navigation between views
type NavigationMsg struct {
	Target ViewType
	Item   MenuItem
}
