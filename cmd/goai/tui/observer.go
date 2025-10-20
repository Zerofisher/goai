package tui

import (
	"context"

	"github.com/Zerofisher/goai/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
)

// Observer implements dispatcher.ToolObserver and sends tool events to the Bubble Tea program
type Observer struct {
	program *tea.Program
}

// NewObserver creates a new observer that sends events to the given Bubble Tea program
func NewObserver(p *tea.Program) *Observer {
	return &Observer{
		program: p,
	}
}

// OnToolEvent implements dispatcher.ToolObserver
// It sends the tool event as a message to the Bubble Tea event loop
func (o *Observer) OnToolEvent(_ context.Context, e types.ToolEvent) {
	if o.program != nil {
		// Non-blocking send to the Bubble Tea message loop
		o.program.Send(ToolEventMsg{Event: e})
	}
}
