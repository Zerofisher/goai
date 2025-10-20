package tui

import (
	"github.com/Zerofisher/goai/pkg/agent"
	"github.com/Zerofisher/goai/pkg/config"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the state of the TUI application
type Model struct {
	// Agent reference for executing queries
	agent *agent.Agent

	// Configuration
	cfg *config.Config

	// Program reference (for sending messages from goroutines)
	program *tea.Program

	// UI Components
	input    textinput.Model // Bottom input field
	chat     viewport.Model  // Left/top viewport for conversation
	tools    viewport.Model  // Right/bottom viewport for tool events
	spinner  spinner.Model   // Loading spinner

	// State
	state struct {
		querying bool   // Whether a query is in progress
		ready    bool   // Whether the UI is ready (window size received)
		width    int    // Terminal width
		height   int    // Terminal height
	}

	// Content buffers
	chatContent  string // Accumulated chat messages
	toolsContent string // Accumulated tool event logs

	// Spinner state
	spinnerLabel string // Current spinner label text
}

// New creates a new TUI model
func New(a *agent.Agent, cfg *config.Config) *Model {
	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = defaultSpinnerStyle()

	m := &Model{
		agent:        a,
		cfg:          cfg,
		input:        ti,
		spinner:      s,
		spinnerLabel: "Ready",
	}

	m.state.ready = false
	m.state.querying = false

	// Initial content
	m.chatContent = "Welcome to GoAI Coder!\n\n"
	m.toolsContent = "Tool events will appear here...\n\n"

	return m
}

// Init initializes the model (Bubble Tea interface)
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}
