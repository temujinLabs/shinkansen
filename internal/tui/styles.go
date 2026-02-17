package tui

import "github.com/charmbracelet/lipgloss"

// Sanzo Wada — Dictionary of Color Combinations Vol. 2
var (
	colorPrimary    = lipgloss.Color("#443b35") // Dull Purplish Black
	colorAccent     = lipgloss.Color("#ae5224") // Burnt Sienna
	colorCTA        = lipgloss.Color("#f37420") // Orange
	colorBackground = lipgloss.Color("#f4deca") // Pale Cinnamon-Pink
	colorSubtle     = lipgloss.Color("#8b7b6f")
	colorWhite      = lipgloss.Color("#ffffff")
)

var (
	// App chrome
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCTA).
			PaddingLeft(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			PaddingLeft(1)

	// Panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorCTA).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			PaddingLeft(1).
			PaddingBottom(1)

	// Issue list
	issueKeyStyle = lipgloss.NewStyle().
			Foreground(colorCTA).
			Bold(true).
			Width(12)

	issueSummaryStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	issueStatusStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Width(14).
				Align(lipgloss.Right)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorAccent).
			Bold(true)

	// Detail view
	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCTA).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	// Help
	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCTA)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Search
	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorCTA).
				Bold(true)

	// Selection indicator
	selectedCheckStyle = lipgloss.NewStyle().
				Foreground(colorCTA).
				Bold(true)
)
