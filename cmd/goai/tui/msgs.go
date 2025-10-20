package tui

import (
	"github.com/Zerofisher/goai/pkg/types"
)

// ToolEventMsg wraps a tool event for the Bubble Tea message loop
type ToolEventMsg struct {
	Event types.ToolEvent
}

// LLMStreamTextMsg contains a chunk of streaming text from the LLM
type LLMStreamTextMsg struct {
	Text string
}

// LLMDoneMsg signals that the LLM has finished generating a response
type LLMDoneMsg struct{}

// ErrorMsg wraps an error for display in the UI
type ErrorMsg struct {
	Err error
}

// QueryMsg triggers a query to the agent
type QueryMsg struct {
	Text string
}
