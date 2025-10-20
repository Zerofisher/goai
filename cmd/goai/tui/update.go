package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case ToolEventMsg:
		return m.handleToolEventMsg(msg)

	case LLMStreamTextMsg:
		return m.handleLLMStreamTextMsg(msg)

	case LLMDoneMsg:
		return m.handleLLMDoneMsg(msg)

	case ErrorMsg:
		return m.handleErrorMsg(msg)

	case QueryMsg:
		return m.handleQueryMsg(msg)
	}

	// Update child components
	var cmd tea.Cmd

	// Update spinner
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	// Update input
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update chat viewport
	m.chat, cmd = m.chat.Update(msg)
	cmds = append(cmds, cmd)

	// Update tools viewport
	m.tools, cmd = m.tools.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		// Trigger query
		if !m.state.querying && m.input.Value() != "" {
			query := m.input.Value()
			m.input.SetValue("")
			return m, func() tea.Msg {
				return QueryMsg{Text: query}
			}
		}

	case "esc":
		// Cancel current operation (if any)
		if m.state.querying {
			m.state.querying = false
			m.spinnerLabel = "Cancelled"
		}
	}

	// Update input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handleWindowSizeMsg handles terminal resize
func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.state.width = msg.Width
	m.state.height = msg.Height
	m.state.ready = true

	// Initialize viewports if not already done
	if m.chat.Width == 0 {
		chatWidth := msg.Width/2 - 4
		toolsWidth := msg.Width/2 - 4
		viewportHeight := msg.Height - 8

		m.chat = viewport.New(chatWidth, viewportHeight)
		m.tools = viewport.New(toolsWidth, viewportHeight)

		m.chat.SetContent(m.chatContent)
		m.tools.SetContent(m.toolsContent)
	}

	return m, nil
}

// handleToolEventMsg handles tool execution events
func (m *Model) handleToolEventMsg(msg ToolEventMsg) (tea.Model, tea.Cmd) {
	e := msg.Event

	var eventText string
	switch e.Type {
	case types.ToolEventStarted:
		eventText = formatToolEvent("started", e.Name, "started")
		m.spinnerLabel = fmt.Sprintf("Running %s...", e.Name)
		m.state.querying = true

	case types.ToolEventSucceeded:
		duration := ""
		if e.Duration != nil {
			duration = fmt.Sprintf(" (%v)", e.Duration.Round(time.Millisecond))
		}
		output := e.Output
		if len(output) > 100 {
			output = output[:100] + "..."
		}
		eventText = formatToolEvent("succeeded", e.Name, fmt.Sprintf("succeeded%s\n  %s", duration, output))
		m.spinnerLabel = "Thinking..."

	case types.ToolEventFailed:
		eventText = formatToolEvent("failed", e.Name, fmt.Sprintf("failed: %s", e.Error))
		m.spinnerLabel = "Thinking..."
	}

	// Append to tools content
	m.toolsContent = appendToContent(m.toolsContent, eventText)
	m.tools.SetContent(m.toolsContent)
	m.tools.GotoBottom()

	return m, nil
}

// handleLLMStreamTextMsg handles streaming text from LLM
func (m *Model) handleLLMStreamTextMsg(msg LLMStreamTextMsg) (tea.Model, tea.Cmd) {
	// Append text to chat content
	m.chatContent += msg.Text
	m.chat.SetContent(m.chatContent)
	m.chat.GotoBottom()

	return m, nil
}

// handleLLMDoneMsg handles completion of LLM response
func (m *Model) handleLLMDoneMsg(_ LLMDoneMsg) (tea.Model, tea.Cmd) {
	m.state.querying = false
	m.spinnerLabel = "Ready"

	// Add newline after completion
	m.chatContent = appendToContent(m.chatContent, "")
	m.chat.SetContent(m.chatContent)

	return m, nil
}

// handleErrorMsg handles error messages
func (m *Model) handleErrorMsg(msg ErrorMsg) (tea.Model, tea.Cmd) {
	errorText := fmt.Sprintf("\n❌ Error: %v\n", msg.Err)
	m.chatContent = appendToContent(m.chatContent, errorText)
	m.chat.SetContent(m.chatContent)
	m.chat.GotoBottom()

	m.state.querying = false
	m.spinnerLabel = "Error"

	return m, nil
}

// handleQueryMsg handles query execution
func (m *Model) handleQueryMsg(msg QueryMsg) (tea.Model, tea.Cmd) {
	// Add user message to chat
	userMsg := fmt.Sprintf("\n👤 You: %s\n", msg.Text)
	m.chatContent = appendToContent(m.chatContent, userMsg)
	m.chat.SetContent(m.chatContent)
	m.chat.GotoBottom()

	// Set querying state
	m.state.querying = true
	m.spinnerLabel = "Thinking..."

	// Start query in background
	return m, func() tea.Msg {
		// Execute query using StreamQuery
		outputChan := make(chan string, 100)
		errChan := make(chan error, 1)

		go func() {
			ctx := context.Background()
			err := m.agent.StreamQuery(ctx, msg.Text, outputChan)
			if err != nil {
				errChan <- err
			}
			close(outputChan)
			close(errChan)
		}()

		// Collect all output
		go func() {
			// Add assistant prefix
			m.program.Send(LLMStreamTextMsg{Text: "\n🤖 Assistant: "})

			// Track if we got any output
			gotOutput := false

			for text := range outputChan {
				gotOutput = true
				m.program.Send(LLMStreamTextMsg{Text: text})
			}

			// Check for errors first
			select {
			case err := <-errChan:
				if err != nil {
					// Send detailed error message
					errMsg := fmt.Sprintf("\n❌ Error: %v", err)
					m.program.Send(LLMStreamTextMsg{Text: errMsg})
					m.program.Send(LLMDoneMsg{})
					return
				}
			default:
			}

			// If no output was received and no error, show diagnostic message
			if !gotOutput {
				m.program.Send(LLMStreamTextMsg{Text: "\n⚠️  No response from LLM.\nPossible issues:\n- API key may be invalid\n- Model not available\n- Network connectivity\n\nCurrent config:\n- Model: " + m.cfg.Model.Name + "\n- Provider: " + m.cfg.Model.Provider})
			}

			// Signal completion
			m.program.Send(LLMDoneMsg{})
		}()

		return nil
	}
}

// SetProgram sets the program reference (needed for StreamQuery callback)
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}
