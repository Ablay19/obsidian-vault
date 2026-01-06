package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// UserModel represents the user management view
type UserModel struct {
	table      table.Model
	styles     Styles
	loading    bool
	lastUpdate time.Time
	users      []UserRow
}

// UserRow represents a user in the table
type UserRow struct {
	ID        string
	Username  string
	Email     string
	CreatedAt string
	LastSeen  string
	Status    string
}

// NewUsers creates a new users model
func NewUsers(styles Styles) UserModel {
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Username", Width: 20},
		{Title: "Email", Width: 25},
		{Title: "Status", Width: 12},
		{Title: "Created", Width: 18},
		{Title: "Last Seen", Width: 18},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	return UserModel{
		table:      t,
		styles:     styles,
		loading:    true,
		lastUpdate: time.Now(),
		users:      []UserRow{},
	}
}

// Init initializes the users model
func (m UserModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		m.fetchUsersCmd(),
	)
}

// Update updates the users model
func (m UserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, nil // Return to main menu
		case "r", "R":
			m.loading = true
			return m, m.fetchUsersCmd()
		}

	case usersMsg:
		m.users = msg.users
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

// View renders the users view
func (m UserModel) View() string {
	content := m.styles.Card.Render(
		m.styles.Header.Render("👥 User Management") + "\n\n" +
			m.renderContent(),
	)

	footer := m.styles.Footer.Render(
		fmt.Sprintf("Last updated: %s | R: Refresh Users | Q: Back to Menu",
			m.lastUpdate.Format("15:04:05"),
		),
	)

	return content + "\n" + footer
}

// renderContent renders the main content
func (m UserModel) renderContent() string {
	if m.loading {
		return m.styles.Loading.Render("🔄 Loading users...")
	}

	if len(m.users) == 0 {
		return m.styles.Muted.Render("No users found.")
	}

	// Create status indicator for each user
	userRows := make([]table.Row, len(m.users))
	for i, user := range m.users {
		status := "🟢 Active"
		if user.Status != "active" {
			status = "🔴 Inactive"
		}

		userRows[i] = table.Row{
			user.ID,
			user.Username,
			user.Email,
			status,
			user.CreatedAt,
			user.LastSeen,
		}
	}

	m.table.SetRows(userRows)
	return m.table.View()
}

// updateTableRows updates the table with user data
func (m UserModel) updateTableRows() {
	rows := make([]table.Row, len(m.users))
	for i, user := range m.users {
		status := "🟢 Active"
		if user.Status != "active" {
			status = "🔴 Inactive"
		}

		rows[i] = table.Row{
			user.ID,
			user.Username,
			user.Email,
			status,
			user.CreatedAt,
			user.LastSeen,
		}
	}

	m.table.SetRows(rows)
}

// fetchUsersCmd creates a command to fetch users
func (m UserModel) fetchUsersCmd() tea.Cmd {
	return func() tea.Msg {
		users, err := m.fetchRealUsers()
		if err != nil {
			return errorMsgMsg{err: err.Error()}
		}
		return usersMsg{users: users}
	}
}

// fetchRealUsers fetches real user data
// In a real implementation, this would call the actual database/API
func (m UserModel) fetchRealUsers() ([]UserRow, error) {
	// Mock data for demonstration
	// TODO: Replace with actual database/API calls
	mockUsers := []UserRow{
		{
			ID:        "1",
			Username:  "admin",
			Email:     "admin@example.com",
			Status:    "active",
			CreatedAt: time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			LastSeen:  time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04"),
		},
		{
			ID:        "2",
			Username:  "user1",
			Email:     "user1@example.com",
			Status:    "active",
			CreatedAt: time.Now().Add(-48 * time.Hour).Format("2006-01-02"),
			LastSeen:  time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04"),
		},
		{
			ID:        "3",
			Username:  "user2",
			Email:     "user2@example.com",
			Status:    "inactive",
			CreatedAt: time.Now().Add(-72 * time.Hour).Format("2006-01-02"),
			LastSeen:  time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04"),
		},
	}

	return mockUsers, nil
}

// Command messages
type usersMsg struct {
	users []UserRow
}
