package message

import (
	"fmt"
	"regexp"
	"strings"
)

// Formatter formats messages for display
type Formatter struct {
	enableColors bool
	codeTheme    string
	wrapLines    bool
	maxWidth     int
}

// NewFormatter creates a new message formatter
func NewFormatter(enableColors bool) *Formatter {
	return &Formatter{
		enableColors: enableColors,
		codeTheme:    "monokai",
		wrapLines:    true,
		maxWidth:     80,
	}
}

// FormatMessage formats a message for display
func (f *Formatter) FormatMessage(role, content string) string {
	if !f.enableColors {
		return fmt.Sprintf("[%s] %s", role, content)
	}

	// Apply colors based on role
	var roleColor, contentColor string
	switch role {
	case "user":
		roleColor = "\033[34m" // Blue
		contentColor = "\033[0m" // Reset
	case "assistant":
		roleColor = "\033[32m" // Green
		contentColor = "\033[0m" // Reset
	case "system":
		roleColor = "\033[33m" // Yellow
		contentColor = "\033[90m" // Gray
	case "tool":
		roleColor = "\033[35m" // Magenta
		contentColor = "\033[90m" // Gray
	default:
		roleColor = "\033[0m"
		contentColor = "\033[0m"
	}

	// Format the content
	formatted := f.formatContent(content)

	return fmt.Sprintf("%s[%s]%s %s%s%s", roleColor, role, "\033[0m", contentColor, formatted, "\033[0m")
}

// formatContent applies formatting to the content
func (f *Formatter) formatContent(content string) string {
	// Format markdown elements
	content = f.formatMarkdown(content)

	// Wrap lines if needed
	if f.wrapLines && f.maxWidth > 0 {
		content = f.wrapText(content, f.maxWidth)
	}

	return content
}

// formatMarkdown applies basic markdown formatting
func (f *Formatter) formatMarkdown(text string) string {
	if !f.enableColors {
		return text
	}

	// Bold: **text** or __text__
	boldRegex := regexp.MustCompile(`(\*\*|__)(.+?)(\*\*|__)`)
	text = boldRegex.ReplaceAllString(text, "\033[1m$2\033[0m")

	// Italic: *text* or _text_
	italicRegex := regexp.MustCompile(`(\*|_)([^\*_]+?)(\*|_)`)
	text = italicRegex.ReplaceAllString(text, "\033[3m$2\033[0m")

	// Code: `code`
	codeRegex := regexp.MustCompile("`([^`]+)`")
	text = codeRegex.ReplaceAllString(text, "\033[38;5;214m$1\033[0m") // Orange

	// Headers: # Header
	headerRegex := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	text = headerRegex.ReplaceAllString(text, "\033[1;4m$2\033[0m") // Bold + Underline

	// Links: [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	text = linkRegex.ReplaceAllString(text, "\033[4;36m$1\033[0m") // Underline + Cyan

	// Code blocks: ```language\ncode\n```
	codeBlockRegex := regexp.MustCompile("(?s)```[^\n]*\n(.*?)```")
	text = codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract the code content
		lines := strings.Split(match, "\n")
		if len(lines) < 2 {
			return match
		}

		var formatted []string
		formatted = append(formatted, "\033[38;5;242m"+lines[0]+"\033[0m") // Gray for ```

		// Apply syntax highlighting to code
		for i := 1; i < len(lines)-1; i++ {
			formatted = append(formatted, f.highlightCode(lines[i]))
		}

		if len(lines) > 1 {
			formatted = append(formatted, "\033[38;5;242m"+lines[len(lines)-1]+"\033[0m") // Gray for ```
		}

		return strings.Join(formatted, "\n")
	})

	// Lists: - item or * item
	listRegex := regexp.MustCompile(`(?m)^(\s*)([-\*])\s+(.+)$`)
	text = listRegex.ReplaceAllString(text, "$1\033[36m$2\033[0m $3") // Cyan bullet

	return text
}

// highlightCode applies basic syntax highlighting to code
func (f *Formatter) highlightCode(line string) string {
	if !f.enableColors {
		return line
	}

	// Keywords (common across languages)
	keywords := []string{
		"if", "else", "for", "while", "return", "func", "function", "def",
		"class", "struct", "interface", "type", "var", "let", "const",
		"import", "package", "from", "export", "module", "require",
		"try", "catch", "finally", "throw", "error", "nil", "null",
		"true", "false", "new", "delete", "this", "self",
	}

	result := line
	for _, keyword := range keywords {
		// Match whole words only
		regex := regexp.MustCompile(`\b` + keyword + `\b`)
		result = regex.ReplaceAllString(result, "\033[38;5;197m"+keyword+"\033[0m") // Pink
	}

	// Strings (simple detection)
	stringRegex := regexp.MustCompile(`"[^"]*"`)
	result = stringRegex.ReplaceAllStringFunc(result, func(match string) string {
		return "\033[38;5;226m" + match + "\033[0m" // Yellow
	})

	stringRegex2 := regexp.MustCompile(`'[^']*'`)
	result = stringRegex2.ReplaceAllStringFunc(result, func(match string) string {
		return "\033[38;5;226m" + match + "\033[0m" // Yellow
	})

	// Comments (simple detection)
	commentRegex := regexp.MustCompile(`//.*$`)
	result = commentRegex.ReplaceAllStringFunc(result, func(match string) string {
		return "\033[38;5;242m" + match + "\033[0m" // Gray
	})

	return result
}

// wrapText wraps text to the specified width
func (f *Formatter) wrapText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var wrapped []string

	for _, line := range lines {
		if len(line) <= width {
			wrapped = append(wrapped, line)
			continue
		}

		// Simple word wrapping
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if len(currentLine)+len(word)+1 > width && currentLine != "" {
				wrapped = append(wrapped, currentLine)
				currentLine = word
			} else {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			}
		}
		if currentLine != "" {
			wrapped = append(wrapped, currentLine)
		}
	}

	return strings.Join(wrapped, "\n")
}

// FormatTable formats data as a table
func (f *Formatter) FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build the table
	var lines []string

	// Header
	var headerLine, separatorLine []string
	for i, h := range headers {
		headerLine = append(headerLine, fmt.Sprintf(" %-*s ", widths[i], h))
		separatorLine = append(separatorLine, strings.Repeat("-", widths[i]+2))
	}

	if f.enableColors {
		lines = append(lines, "\033[1m"+strings.Join(headerLine, "|")+"\033[0m")
	} else {
		lines = append(lines, strings.Join(headerLine, "|"))
	}
	lines = append(lines, strings.Join(separatorLine, "+"))

	// Rows
	for _, row := range rows {
		var rowLine []string
		for i, cell := range row {
			if i < len(widths) {
				rowLine = append(rowLine, fmt.Sprintf(" %-*s ", widths[i], cell))
			}
		}
		lines = append(lines, strings.Join(rowLine, "|"))
	}

	return strings.Join(lines, "\n")
}

// SetEnableColors sets whether colors are enabled
func (f *Formatter) SetEnableColors(enable bool) {
	f.enableColors = enable
}

// SetWrapLines sets whether to wrap lines
func (f *Formatter) SetWrapLines(wrap bool) {
	f.wrapLines = wrap
}

// SetMaxWidth sets the maximum line width for wrapping
func (f *Formatter) SetMaxWidth(width int) {
	f.maxWidth = width
}