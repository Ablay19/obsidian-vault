package views

import (
	"github.com/charmbracelet/lipgloss"
)

// ColorPalette defines the application color scheme
type ColorPalette struct {
	Primary       lipgloss.Color
	Secondary     lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	Info          lipgloss.Color
	Muted         lipgloss.Color
	Background    lipgloss.Color
	Surface       lipgloss.Color
	Border        lipgloss.Color
	Text          lipgloss.Color
	TextSecondary lipgloss.Color
}

// DefaultPalette returns the default color palette
func DefaultPalette() ColorPalette {
	return ColorPalette{
		Primary:       lipgloss.Color("63"),  // Blue
		Secondary:     lipgloss.Color("99"),  // Purple
		Success:       lipgloss.Color("35"),  // Green
		Warning:       lipgloss.Color("208"), // Orange
		Error:         lipgloss.Color("124"), // Red
		Info:          lipgloss.Color("39"),  // Cyan
		Muted:         lipgloss.Color("245"), // Gray
		Background:    lipgloss.Color("16"),  // Black
		Surface:       lipgloss.Color("235"), // Dark gray
		Border:        lipgloss.Color("240"), // Gray
		Text:          lipgloss.Color("255"), // White
		TextSecondary: lipgloss.Color("241"), // Light gray
	}
}

// Styles contains all styled components
type Styles struct {
	// Common styles
	Header    lipgloss.Style
	SubHeader lipgloss.Style
	Body      lipgloss.Style
	Footer    lipgloss.Style

	// Interactive elements
	Button       lipgloss.Style
	ButtonActive lipgloss.Style

	// Status indicators
	Loading lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Container styles
	Card   lipgloss.Style
	Border lipgloss.Style
	Margin lipgloss.Style

	// Text styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
	Muted    lipgloss.Style

	// Table styles
	Table       lipgloss.Style
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableBorder lipgloss.Style

	// List styles
	List       lipgloss.Style
	ListItem   lipgloss.Style
	ListActive lipgloss.Style

	// Navigation styles
	Help       lipgloss.Style
	Pagination lipgloss.Style

	// Menu styles
	MenuTitle    lipgloss.Style
	MenuActive   lipgloss.Style
	MenuInactive lipgloss.Style
}

// NewStyles creates a new style set with the given color palette
func NewStyles(palette ColorPalette) Styles {
	return Styles{
		// Common styles
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Primary).
			Padding(0, 1).
			MarginBottom(1),

		SubHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.TextSecondary).
			Padding(0, 1).
			MarginBottom(1),

		Body: lipgloss.NewStyle().
			Foreground(palette.Text).
			Padding(1),

		Footer: lipgloss.NewStyle().
			Foreground(palette.TextSecondary).
			Italic(true).
			MarginTop(1),

		// Interactive elements
		Button: lipgloss.NewStyle().
			Foreground(palette.Background).
			Background(palette.Primary).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder(), true, true, true, true),

		ButtonActive: lipgloss.NewStyle().
			Foreground(palette.Background).
			Background(palette.Secondary).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder(), true, true, true, true),

		// Status indicators
		Loading: lipgloss.NewStyle().
			Foreground(palette.Info).
			Italic(true),

		Success: lipgloss.NewStyle().
			Foreground(palette.Success).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(palette.Error).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(palette.Warning).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(palette.Info),

		// Container styles
		Card: lipgloss.NewStyle().
			Background(palette.Surface).
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			BorderForeground(palette.Border).
			Padding(1, 2).
			Margin(1),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			BorderForeground(palette.Border),

		Margin: lipgloss.NewStyle().
			Margin(1),

		// Text styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Text).
			Underline(true),

		Subtitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.TextSecondary),

		Label: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.TextSecondary),

		Value: lipgloss.NewStyle().
			Foreground(palette.Text),

		Muted: lipgloss.NewStyle().
			Foreground(palette.TextSecondary).
			Italic(true),

		// Table styles
		Table: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(palette.Border),

		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Primary).
			Padding(0, 1),

		TableRow: lipgloss.NewStyle().
			Foreground(palette.Text).
			Padding(0, 1),

		TableBorder: lipgloss.NewStyle().
			Foreground(palette.Border),

		// List styles
		List: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			BorderForeground(palette.Border),

		ListItem: lipgloss.NewStyle().
			Foreground(palette.Text).
			PaddingLeft(2).
			PaddingRight(1),

		ListActive: lipgloss.NewStyle().
			Foreground(palette.Primary).
			Background(palette.Surface).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(1),

		// Navigation styles
		Help: lipgloss.NewStyle().
			Foreground(palette.TextSecondary).
			Italic(true).
			MarginTop(1),

		Pagination: lipgloss.NewStyle().
			Foreground(palette.TextSecondary).
			PaddingLeft(1),

		// Menu styles
		MenuTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(palette.Primary).
			Padding(0, 1).
			MarginBottom(1),

		MenuActive: lipgloss.NewStyle().
			Foreground(palette.Primary).
			Background(palette.Surface).
			Bold(true).
			PaddingLeft(1),

		MenuInactive: lipgloss.NewStyle().
			Foreground(palette.TextSecondary).
			PaddingLeft(3),
	}
}
