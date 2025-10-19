package bash

import (
	"regexp"
	"strings"
	"unicode"
)

// OutputProcessor processes command output for better readability.
type OutputProcessor struct {
	ansiRegex      *regexp.Regexp
	maxLineLength  int
	maxLines       int
}

// NewOutputProcessor creates a new output processor.
func NewOutputProcessor() *OutputProcessor {
	return &OutputProcessor{
		// Regex to match ANSI escape codes
		ansiRegex:     regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`),
		maxLineLength: 1000, // Default max line length
		maxLines:      5000, // Default max lines
	}
}

// ProcessOutput processes the command output.
func (p *OutputProcessor) ProcessOutput(output string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = p.maxLines
	}

	// Remove ANSI codes
	cleaned := p.RemoveANSICodes(output)

	// Split into lines
	lines := strings.Split(cleaned, "\n")

	// Truncate if needed
	if len(lines) > maxLines {
		lines = p.truncateOutput(lines, maxLines)
	}

	// Process each line
	var processedLines []string
	for _, line := range lines {
		processed := p.processLine(line)
		if processed != "" || len(processedLines) == 0 {
			processedLines = append(processedLines, processed)
		}
	}

	// Remove excessive blank lines
	processedLines = p.removeExcessiveBlankLines(processedLines)

	return strings.Join(processedLines, "\n")
}

// RemoveANSICodes removes ANSI escape codes from the output.
func (p *OutputProcessor) RemoveANSICodes(output string) string {
	return p.ansiRegex.ReplaceAllString(output, "")
}

// processLine processes a single line of output.
func (p *OutputProcessor) processLine(line string) string {
	// Trim trailing whitespace
	line = strings.TrimRightFunc(line, unicode.IsSpace)

	// Truncate very long lines
	if len(line) > p.maxLineLength {
		line = line[:p.maxLineLength] + "... [truncated]"
	}

	// Remove control characters except tab
	line = p.removeControlCharacters(line)

	return line
}

// removeControlCharacters removes control characters except tab.
func (p *OutputProcessor) removeControlCharacters(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r == '\t' || (r >= 32 && r != 127) || r == '\n' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// truncateOutput truncates the output to maxLines, keeping beginning and end.
func (p *OutputProcessor) truncateOutput(lines []string, maxLines int) []string {
	if len(lines) <= maxLines {
		return lines
	}

	// Keep first and last portions
	keepStart := maxLines * 3 / 4
	keepEnd := maxLines - keepStart - 1

	result := make([]string, 0, maxLines)
	result = append(result, lines[:keepStart]...)
	result = append(result, "\n... [truncated middle portion] ...\n")
	result = append(result, lines[len(lines)-keepEnd:]...)

	return result
}

// removeExcessiveBlankLines removes multiple consecutive blank lines.
func (p *OutputProcessor) removeExcessiveBlankLines(lines []string) []string {
	var result []string
	blankCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				result = append(result, line)
			}
		} else {
			blankCount = 0
			result = append(result, line)
		}
	}

	return result
}

// FormatError formats an error message for display.
func (p *OutputProcessor) FormatError(err error, command string) string {
	var result strings.Builder

	result.WriteString("❌ Command failed\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("Command: ")
	result.WriteString(p.truncateCommand(command))
	result.WriteString("\n")
	result.WriteString("Error: ")
	result.WriteString(err.Error())
	result.WriteString("\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return result.String()
}

// FormatSuccess formats a success message for display.
func (p *OutputProcessor) FormatSuccess(output string, command string) string {
	var result strings.Builder

	result.WriteString("✅ Command completed successfully\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	result.WriteString("Command: ")
	result.WriteString(p.truncateCommand(command))
	result.WriteString("\n")

	if output != "" {
		result.WriteString("Output:\n")
		result.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			result.WriteString("\n")
		}
	} else {
		result.WriteString("(no output)\n")
	}

	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return result.String()
}

// truncateCommand truncates a long command for display.
func (p *OutputProcessor) truncateCommand(command string) string {
	maxLen := 100
	if len(command) <= maxLen {
		return command
	}
	return command[:maxLen-3] + "..."
}

// HighlightSyntax adds simple syntax highlighting for common patterns.
func (p *OutputProcessor) HighlightSyntax(output string) string {
	// This is a simplified version - in production, you might want
	// to use a proper syntax highlighting library

	patterns := map[string]string{
		// Error patterns
		`(?i)(error|err|fatal|failed|failure)`:   "❌ $1",
		`(?i)(warning|warn)`:                      "⚠️  $1",
		`(?i)(success|succeeded|ok|passed|done)`: "✅ $1",

		// File paths
		`(/[a-zA-Z0-9_.-]+)+`: "📁 $0",

		// IP addresses
		`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`: "🌐 $0",

		// URLs
		`https?://[^\s]+`: "🔗 $0",
	}

	result := output
	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, replacement)
	}

	return result
}

// ExtractErrors attempts to extract error messages from output.
func (p *OutputProcessor) ExtractErrors(output string) []string {
	var errors []string

	// Common error patterns
	errorPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)error:\s*(.+)`),
		regexp.MustCompile(`(?i)fatal:\s*(.+)`),
		regexp.MustCompile(`(?i)failed:\s*(.+)`),
		regexp.MustCompile(`(?i)cannot\s+(.+)`),
		regexp.MustCompile(`(?i)unable\s+to\s+(.+)`),
		regexp.MustCompile(`(?i)permission\s+denied`),
		regexp.MustCompile(`(?i)no\s+such\s+file\s+or\s+directory`),
		regexp.MustCompile(`(?i)command\s+not\s+found`),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, pattern := range errorPatterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				errors = append(errors, strings.TrimSpace(line))
				break
			}
		}
	}

	return errors
}

// SummarizeOutput creates a summary of long output.
func (p *OutputProcessor) SummarizeOutput(output string) string {
	lines := strings.Split(output, "\n")

	if len(lines) <= 20 {
		return output
	}

	var summary strings.Builder

	// Add line count
	summary.WriteString("📊 Output Summary\n")
	summary.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	summary.WriteString("Total lines: ")
	summary.WriteString(string(rune(len(lines))))
	summary.WriteString("\n")

	// Extract errors if any
	errors := p.ExtractErrors(output)
	if len(errors) > 0 {
		summary.WriteString("\n❌ Errors found:\n")
		for i, err := range errors {
			if i >= 5 {
				summary.WriteString("... and ")
				summary.WriteString(string(rune(len(errors) - 5)))
				summary.WriteString(" more errors\n")
				break
			}
			summary.WriteString("  • ")
			summary.WriteString(err)
			summary.WriteString("\n")
		}
	}

	// Show first few lines
	summary.WriteString("\n📝 First 5 lines:\n")
	for i := 0; i < 5 && i < len(lines); i++ {
		summary.WriteString("  ")
		summary.WriteString(p.truncateCommand(lines[i]))
		summary.WriteString("\n")
	}

	// Show last few lines
	summary.WriteString("\n📝 Last 5 lines:\n")
	start := len(lines) - 5
	if start < 0 {
		start = 0
	}
	for i := start; i < len(lines); i++ {
		summary.WriteString("  ")
		summary.WriteString(p.truncateCommand(lines[i]))
		summary.WriteString("\n")
	}

	summary.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return summary.String()
}