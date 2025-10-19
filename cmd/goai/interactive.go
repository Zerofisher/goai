package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/Zerofisher/goai/pkg/agent"
	"github.com/Zerofisher/goai/pkg/config"
)

// Color constants for terminal output
const (
	resetColor   = "\033[0m"
	primaryColor = "\033[36m"  // Cyan
	successColor = "\033[32m"  // Green
	errorColor   = "\033[31m"  // Red
	infoColor    = "\033[33m"  // Yellow
	dimColor     = "\033[90m"  // Dim gray
)

// InteractiveSession manages the interactive user interface.
type InteractiveSession struct {
	config  *config.Config
	spinner *Spinner
	agent   *agent.Agent
}

// NewInteractiveSession creates a new interactive session.
func NewInteractiveSession(cfg *config.Config, a *agent.Agent) *InteractiveSession {
	return &InteractiveSession{
		config:  cfg,
		spinner: NewSpinner(),
		agent:   a,
	}
}

// PrintWelcome displays the welcome message.
func (s *InteractiveSession) PrintWelcome() {
	clearScreen()
	displayBanner()
	fmt.Printf("%sReady to help! Type 'exit' to quit, 'clear' to clear screen, or 'help' for more commands.%s\n\n", infoColor, resetColor)
}

// PrintPrompt displays the input prompt.
func (s *InteractiveSession) PrintPrompt() {
	fmt.Printf("%s> %s", primaryColor, resetColor)
}

// PrintError displays an error message.
func (s *InteractiveSession) PrintError(err error) {
	fmt.Printf("%sError: %v%s\n", errorColor, err, resetColor)
}

// PrintSuccess displays a success message.
func (s *InteractiveSession) PrintSuccess(message string) {
	fmt.Printf("%s%s%s\n", successColor, message, resetColor)
}

// PrintInfo displays an informational message.
func (s *InteractiveSession) PrintInfo(message string) {
	fmt.Printf("%s%s%s\n", infoColor, message, resetColor)
}

// StartSpinner starts the loading spinner.
func (s *InteractiveSession) StartSpinner(message string) {
	if s.config.Output.ShowSpinner {
		s.spinner.Start(message)
	}
}

// StopSpinner stops the loading spinner.
func (s *InteractiveSession) StopSpinner() {
	s.spinner.Stop()
}

// FormatResponse formats the agent's response for display.
func (s *InteractiveSession) FormatResponse(response string) string {
	return formatMarkdown(response)
}

// HandleExit performs cleanup and exits the application.
func (s *InteractiveSession) HandleExit() {
	fmt.Printf("%sGoodbye!%s\n", primaryColor, resetColor)
	os.Exit(0)
}

// runInteractiveLoop runs the main interactive loop
func runInteractiveLoop(ctx context.Context, a *agent.Agent) {
	session := NewInteractiveSession(a.GetConfig(), a)

	// Create readline instance with UTF-8 support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s> %s", primaryColor, resetColor),
		HistoryFile:     "/tmp/.goai_history",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistoryLimit:    1000,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
	})
	if err != nil {
		fmt.Printf("Error creating readline: %v\n", err)
		return
	}
	defer func() {
		_ = rl.Close() // Non-critical error, can be ignored in cleanup
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := rl.Readline()
			if err != nil {
				// Handle EOF or interrupt
				session.HandleExit()
			}

			input := strings.TrimSpace(line)
			if input == "" {
				continue
			}

			// Handle special commands
			if handleSpecialCommand(input, session, a) {
				continue
			}

			// Process regular query with timeout
			session.StartSpinner("Thinking...")

			// Create a context with timeout for this query (120 seconds)
			queryCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
			response, err := a.Query(queryCtx, input)
			cancel() // Clean up

			session.StopSpinner()

			if err != nil {
				// Check if it's a timeout error
				if queryCtx.Err() == context.DeadlineExceeded {
					session.PrintError(fmt.Errorf("request timeout after 120 seconds. Please try again or use a simpler query"))
				} else {
					session.PrintError(err)
				}
			} else {
				fmt.Println(session.FormatResponse(response))
			}
		}
	}
}

// handleSpecialCommand handles special commands and returns true if handled
func handleSpecialCommand(input string, session *InteractiveSession, a *agent.Agent) bool {
	lowered := strings.ToLower(strings.TrimSpace(input))

	// Remove leading slash if present
	lowered = strings.TrimPrefix(lowered, "/")

	switch lowered {
	case "exit", "quit", "bye":
		session.HandleExit()
		return true

	case "clear", "cls":
		clearScreen()
		return true

	case "help", "?", "h":
		printHelp()
		return true

	case "stats", "s":
		printStats(a)
		return true

	case "reset", "r":
		a.Reset()
		session.PrintSuccess("Agent state has been reset.")
		return true

	default:
		return false
	}
}

// printStats displays agent statistics
func printStats(a *agent.Agent) {
	stats := a.GetStats()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Agent Statistics")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Messages:     %d\n", stats.MessageCount)
	fmt.Printf("Tokens Used:  ~%d\n", stats.TokenCount)
	fmt.Printf("Tool Calls:   %d\n", stats.ToolCallCount)
	fmt.Printf("Errors:       %d\n", stats.ErrorCount)
	fmt.Printf("Total Rounds: %d\n", stats.TotalRounds)
	fmt.Printf("Uptime:       %s\n", stats.Uptime)
	fmt.Printf("Summary:      %s\n", stats.MessageSummary)
	fmt.Println(strings.Repeat("-", 50) + "\n")
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// displayBanner displays the application banner
func displayBanner() {
	banner := `
╔══════════════════════════════════════════════════════╗
║                    GoAI Coder v0.1.0                 ║
║        Your Intelligent Programming Assistant        ║
╚══════════════════════════════════════════════════════╝
`
	fmt.Printf("%s%s%s\n", primaryColor, banner, resetColor)
}

// formatMarkdown performs basic markdown formatting for terminal output
func formatMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var formatted []string

	for _, line := range lines {
		// Format headers
		if strings.HasPrefix(line, "### ") {
			line = fmt.Sprintf("%s%s%s", primaryColor, line, resetColor)
		} else if strings.HasPrefix(line, "## ") {
			line = fmt.Sprintf("%s%s%s", primaryColor, strings.ToUpper(strings.TrimPrefix(line, "## ")), resetColor)
		} else if strings.HasPrefix(line, "# ") {
			line = fmt.Sprintf("%s%s%s", primaryColor, strings.ToUpper(strings.TrimPrefix(line, "# ")), resetColor)
		}

		// Format code blocks
		if strings.HasPrefix(line, "```") {
			line = fmt.Sprintf("%s%s%s", dimColor, line, resetColor)
		}

		// Format bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = fmt.Sprintf("%s•%s%s", successColor, resetColor, line[1:])
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}