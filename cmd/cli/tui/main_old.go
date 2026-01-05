

package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"obsidian-automation/cmd/cli/tui/views"
)

type view int

const (
	menuView view = iota
	statusView
	aiProvidersView
	usersView
)

type mainModel struct {
	currentView view
	menu        views.Model
	status      views.StatusModel
	aiProviders views.AIProvidersModel
	users       views.UsersModel
	quitting    bool
}

func initialModel() mainModel {
	return mainModel{
		currentView: menuView,
		menu:        views.New(),
		status:      views.NewStatus(),
		aiProviders: views.NewAIProviders(),
		users:       views.NewUsers(),
	}
}

func (m mainModel) Init() tea.Cmd {
	return m.menu.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q", "esc":
			if m.currentView != menuView {
				m.currentView = menuView
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	var currentModel tea.Model

	switch m.currentView {
	case menuView:
		currentModel, cmd = m.menu.Update(msg)
		m.menu = currentModel.(views.Model)
		if m.menu.Choice != "" {
			switch m.menu.Choice {
			case "Bot Status":
				m.currentView = statusView
			case "AI Providers":
				m.currentView = aiProvidersView
			case "User Management":
				m.currentView = usersView
			}
			m.menu.Choice = "" // Reset choice
		}
		if m.menu.Quitting {
			m.quitting = true
			return m, tea.Quit
		}

	case statusView:
		currentModel, cmd = m.status.Update(msg)
		m.status = currentModel.(views.StatusModel)

	case aiProvidersView:
		currentModel, cmd = m.aiProviders.Update(msg)
		m.aiProviders = currentModel.(views.AIProvidersModel)

	case usersView:
		currentModel, cmd = m.users.Update(msg)
		m.users = currentModel.(views.UsersModel)
	}

	return m, cmd
}

func (m mainModel) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	switch m.currentView {
	case statusView:
		return m.status.View()
	case aiProvidersView:
		return m.aiProviders.View()
	case usersView:
		return m.users.View()
	default:
		return m.menu.View()
	}
}

func Run() {
	m := initialModel()
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
