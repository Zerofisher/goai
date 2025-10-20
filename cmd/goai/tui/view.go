package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state of the TUI
func (m *Model) View() string {
	if !m.state.ready {
		return "\n  Initializing..."
	}

	// Layout dimensions
	width := m.state.width
	height := m.state.height

	// Calculate viewport dimensions
	// Layout: [Chat | Tools] (side by side)
	//         [Input]
	//         [Status]

	chatWidth := width / 2 - 4
	toolsWidth := width / 2 - 4
	viewportHeight := height - 8 // Leave room for input and status

	// Update viewport sizes
	m.chat.Width = chatWidth
	m.chat.Height = viewportHeight
	m.tools.Width = toolsWidth
	m.tools.Height = viewportHeight

	// Render viewports
	chatView := viewportStyle.
		Width(chatWidth).
		Height(viewportHeight).
		Render(m.chat.View())

	toolsView := viewportStyle.
		Width(toolsWidth).
		Height(viewportHeight).
		Render(m.tools.View())

	// Combine chat and tools side by side
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		chatView,
		toolsView,
	)

	// Render input
	inputView := inputStyle.
		Width(width - 4).
		Render(m.input.View())

	// Render status bar
	statusText := ""
	if m.state.querying {
		statusText = fmt.Sprintf("%s %s", m.spinner.View(), m.spinnerLabel)
	} else {
		statusText = "Ready • Press Ctrl+C to quit"
	}
	statusView := statusBarStyle.Width(width - 2).Render(statusText)

	// Combine all views vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("GoAI Coder - Chat"),
		mainView,
		inputView,
		statusView,
	)
}

// Helper function to render tool event
func formatToolEvent(eventType, name, message string) string {
	var style lipgloss.Style
	var prefix string

	switch eventType {
	case "started":
		style = toolStartedStyle
		prefix = "▶"
	case "succeeded":
		style = toolSucceededStyle
		prefix = "✓"
	case "failed":
		style = toolFailedStyle
		prefix = "✗"
	default:
		style = lipgloss.NewStyle()
		prefix = "●"
	}

	return style.Render(fmt.Sprintf("%s [%s] %s", prefix, name, message))
}

// appendToContent appends text to a content buffer with proper newlines
func appendToContent(content, newText string) string {
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + newText
}
