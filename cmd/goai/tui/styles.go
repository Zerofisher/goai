package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	colorPrimary   = lipgloss.Color("#00D7AF")
	colorSecondary = lipgloss.Color("#7D56F4")
	colorError     = lipgloss.Color("#FF5F87")
	colorSubtle    = lipgloss.Color("#6C7086")
	colorBorder    = lipgloss.Color("#45475A")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	// Viewport styles
	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// Input styles
	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Spinner styles
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// Status bar style
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Padding(0, 1)

	// Tool event styles
	toolStartedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500"))

	toolSucceededStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	toolFailedStyle = lipgloss.NewStyle().
				Foreground(colorError)
)

// defaultSpinnerStyle returns the default spinner style
func defaultSpinnerStyle() lipgloss.Style {
	return spinnerStyle
}
