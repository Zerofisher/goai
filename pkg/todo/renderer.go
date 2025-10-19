package todo

import (
	"fmt"
	"strings"
)

// ANSI color codes for terminal output
const (
	// ResetColor resets all formatting
	ResetColor = "\x1b[0m"

	// TodoPendingColor is gray for pending tasks
	TodoPendingColor = "\x1b[38;2;176;176;176m"

	// TodoProgressColor is blue for in-progress tasks
	TodoProgressColor = "\x1b[38;2;120;200;255m"

	// TodoCompletedColor is green for completed tasks
	TodoCompletedColor = "\x1b[38;2;34;139;34m"

	// StrikethroughStyle adds strikethrough formatting
	StrikethroughStyle = "\x1b[9m"
)

// Renderer handles the visual formatting of todo lists.
type Renderer struct {
	enableColors bool
	showStats    bool
}

// NewRenderer creates a new todo renderer.
func NewRenderer(enableColors bool) *Renderer {
	return &Renderer{
		enableColors: enableColors,
		showStats:    true,
	}
}

// SetShowStats configures whether to show statistics after the todo list.
func (r *Renderer) SetShowStats(show bool) {
	r.showStats = show
}

// Render formats a list of todo items into a string representation.
func (r *Renderer) Render(items []TodoItem) string {
	if len(items) == 0 {
		return r.formatEmpty()
	}

	var lines []string
	for _, item := range items {
		lines = append(lines, r.renderItem(item))
	}

	result := strings.Join(lines, "\n")

	// Add statistics if enabled
	if r.showStats {
		stats := r.calculateStats(items)
		result += "\n\n" + r.renderStats(stats)
	}

	return result
}

// RenderWithStats renders the todo list with statistics.
func (r *Renderer) RenderWithStats(items []TodoItem, stats Stats) string {
	if len(items) == 0 {
		return r.formatEmpty()
	}

	var lines []string
	for _, item := range items {
		lines = append(lines, r.renderItem(item))
	}

	result := strings.Join(lines, "\n")
	result += "\n\n" + r.renderStats(stats)

	return result
}

// renderItem formats a single todo item.
func (r *Renderer) renderItem(item TodoItem) string {
	mark := r.getStatusMark(item.Status)
	text := fmt.Sprintf("%s %s", mark, item.Content)

	// Show active form for in-progress items if different from content
	if item.Status == StatusInProgress && item.ActiveForm != "" && item.ActiveForm != item.Content {
		text = fmt.Sprintf("%s %s (%s)", mark, item.Content, item.ActiveForm)
	}

	if !r.enableColors {
		return text
	}

	// Apply color and styling based on status
	switch item.Status {
	case StatusCompleted:
		// Green with strikethrough for completed items
		return fmt.Sprintf("%s%s%s%s", TodoCompletedColor, StrikethroughStyle, text, ResetColor)
	case StatusInProgress:
		// Blue for in-progress items
		return fmt.Sprintf("%s%s%s", TodoProgressColor, text, ResetColor)
	case StatusPending:
		// Gray for pending items
		return fmt.Sprintf("%s%s%s", TodoPendingColor, text, ResetColor)
	default:
		return text
	}
}

// getStatusMark returns the checkbox mark for a given status.
func (r *Renderer) getStatusMark(status Status) string {
	switch status {
	case StatusCompleted:
		return "☒"
	case StatusInProgress:
		return "☐"
	case StatusPending:
		return "☐"
	default:
		return "☐"
	}
}

// formatEmpty returns the message for an empty todo list.
func (r *Renderer) formatEmpty() string {
	text := "☐ No todos yet"
	if r.enableColors {
		return fmt.Sprintf("%s%s%s", TodoPendingColor, text, ResetColor)
	}
	return text
}

// calculateStats computes statistics from a list of todo items.
func (r *Renderer) calculateStats(items []TodoItem) Stats {
	stats := Stats{
		Total: len(items),
	}

	for _, item := range items {
		switch item.Status {
		case StatusCompleted:
			stats.Completed++
		case StatusInProgress:
			stats.InProgress++
		case StatusPending:
			stats.Pending++
		}
	}

	return stats
}

// renderStats formats statistics into a string.
func (r *Renderer) renderStats(stats Stats) string {
	if stats.Total == 0 {
		return "No todos have been created."
	}

	parts := []string{
		fmt.Sprintf("Total: %d", stats.Total),
		fmt.Sprintf("Completed: %d", stats.Completed),
		fmt.Sprintf("In Progress: %d", stats.InProgress),
		fmt.Sprintf("Pending: %d", stats.Pending),
	}

	// Calculate completion percentage
	if stats.Total > 0 {
		percentage := (stats.Completed * 100) / stats.Total
		parts = append(parts, fmt.Sprintf("Completion: %d%%", percentage))
	}

	return "📊 " + strings.Join(parts, " | ")
}

// RenderCompact returns a compact single-line summary of the todo list.
func (r *Renderer) RenderCompact(items []TodoItem) string {
	stats := r.calculateStats(items)

	if stats.Total == 0 {
		return "No todos"
	}

	// Show a progress bar
	progressBar := r.renderProgressBar(stats)

	return fmt.Sprintf("[%s] %d/%d completed", progressBar, stats.Completed, stats.Total)
}

// renderProgressBar creates a visual progress bar.
func (r *Renderer) renderProgressBar(stats Stats) string {
	if stats.Total == 0 {
		return ""
	}

	const barWidth = 10
	completed := (stats.Completed * barWidth) / stats.Total
	inProgress := (stats.InProgress * barWidth) / stats.Total

	// Ensure we show at least one character for in-progress if there are any
	if stats.InProgress > 0 && inProgress == 0 {
		inProgress = 1
	}

	pending := barWidth - completed - inProgress

	bar := strings.Repeat("█", completed) +
		strings.Repeat("▒", inProgress) +
		strings.Repeat("░", pending)

	if r.enableColors {
		// Color the bar segments
		coloredBar := ""
		if completed > 0 {
			coloredBar += TodoCompletedColor + strings.Repeat("█", completed) + ResetColor
		}
		if inProgress > 0 {
			coloredBar += TodoProgressColor + strings.Repeat("▒", inProgress) + ResetColor
		}
		if pending > 0 {
			coloredBar += TodoPendingColor + strings.Repeat("░", pending) + ResetColor
		}
		return coloredBar
	}

	return bar
}

// RenderMarkdown formats the todo list as Markdown.
func (r *Renderer) RenderMarkdown(items []TodoItem) string {
	if len(items) == 0 {
		return "- [ ] No todos yet\n"
	}

	var lines []string
	lines = append(lines, "## Todo List\n")

	for _, item := range items {
		checkbox := "- [ ]"
		if item.Status == StatusCompleted {
			checkbox = "- [x]"
		}

		line := fmt.Sprintf("%s %s", checkbox, item.Content)

		// Add status indicator for in-progress
		if item.Status == StatusInProgress {
			line += " *(in progress)*"
		}

		lines = append(lines, line)
	}

	// Add statistics
	stats := r.calculateStats(items)
	lines = append(lines, "")
	lines = append(lines, "### Statistics")
	lines = append(lines, fmt.Sprintf("- Total: %d", stats.Total))
	lines = append(lines, fmt.Sprintf("- Completed: %d", stats.Completed))
	lines = append(lines, fmt.Sprintf("- In Progress: %d", stats.InProgress))
	lines = append(lines, fmt.Sprintf("- Pending: %d", stats.Pending))

	return strings.Join(lines, "\n")
}