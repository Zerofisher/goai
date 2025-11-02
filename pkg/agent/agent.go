package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zerofisher/goai/pkg/config"
	"github.com/Zerofisher/goai/pkg/dispatcher"
	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/message"
	"github.com/Zerofisher/goai/pkg/prompt"
	"github.com/Zerofisher/goai/pkg/types"
)

// Agent represents the main agent that manages interactions between user, LLM, and tools
type Agent struct {
	client        llm.Client
	messages      *message.Manager
	dispatcher    *dispatcher.Dispatcher
	config        *config.Config
	state         *State
	context       *Context
	promptManager *prompt.Manager
	mu            sync.RWMutex
}

// NewAgent creates a new agent with the given configuration
func NewAgent(cfg *config.Config) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create LLM client
	clientConfig := llm.ClientConfig{
		Provider:    cfg.Model.Provider,
		APIKey:      cfg.Model.APIKey,
		BaseURL:     cfg.Model.BaseURL,
		Model:       cfg.Model.Name,
		MaxTokens:   cfg.Model.MaxTokens,
		Temperature: 0.7,
		Timeout:     time.Duration(cfg.Model.Timeout) * time.Second,
	}

	client, err := llm.CreateClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create message manager
	messageManager := message.NewManager(cfg.Model.MaxTokens)

	// Create dispatcher
	toolDispatcher := dispatcher.New(cfg.WorkDir)

	// Create context
	agentContext := NewContext(cfg.WorkDir)

	// Create prompt manager
	promptMgr := prompt.NewManager(cfg, agentContext)

	// Create agent
	agent := &Agent{
		client:        client,
		messages:      messageManager,
		dispatcher:    toolDispatcher,
		config:        cfg,
		state:         NewState(),
		context:       agentContext,
		promptManager: promptMgr,
	}

	// Set up dynamic tool list provider for prompt manager
	promptMgr.SetToolListProvider(func() []string {
		tools := agent.dispatcher.ListTools()
		toolNames := make([]string, len(tools))
		for i, tool := range tools {
			toolNames[i] = tool.Name()
		}
		return toolNames
	})

	// Initialize system prompt
	if err := agent.initializeSystemPrompt(); err != nil {
		return nil, fmt.Errorf("failed to initialize system prompt: %w", err)
	}

	return agent, nil
}

// Query processes a user query and returns the response
func (a *Agent) Query(ctx context.Context, input string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update state
	a.state.IncrementRound()
	a.state.SetProcessing(true)
	defer a.state.SetProcessing(false)

	// Add user message
	if err := a.messages.Add(types.NewTextMessage("user", input)); err != nil {
		return "", fmt.Errorf("failed to add user message: %w", err)
	}

	// Create LLM request
	req := a.buildMessageRequest()

	// Send to LLM
	resp, err := a.client.CreateMessage(ctx, req)
	if err != nil {
		a.state.RecordError(err)
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	// Add assistant message to history
	if err := a.messages.Add(resp.Message); err != nil {
		return "", fmt.Errorf("failed to add assistant message: %w", err)
	}

	// Support multiple rounds of tool calls
	maxRounds := 10 // Prevent infinite loops
	currentRound := 0

	for resp.Message.HasToolUse() && currentRound < maxRounds {
		currentRound++
		toolUses := resp.Message.GetToolUses()

		// Execute tools
		results := a.ProcessToolCalls(toolUses)

		// Add tool results to messages
		for _, result := range results {
			if err := a.messages.Add(types.NewToolResultMessage(&result)); err != nil {
				return "", fmt.Errorf("failed to add tool result: %w", err)
			}
		}

		// Get next response from LLM
		req = a.buildMessageRequest()
		resp, err = a.client.CreateMessage(ctx, req)
		if err != nil {
			a.state.RecordError(err)
			return "", fmt.Errorf("LLM request failed in round %d: %w", currentRound, err)
		}

		// Add assistant message
		if err := a.messages.Add(resp.Message); err != nil {
			return "", fmt.Errorf("failed to add assistant message: %w", err)
		}
	}

	// Extract final text response
	return resp.Message.GetText(), nil
}

// ProcessToolCalls processes a list of tool calls and returns results
func (a *Agent) ProcessToolCalls(calls []*types.ToolUse) []types.ToolResult {
	results := make([]types.ToolResult, len(calls))

	// Process tools in parallel if multiple calls
	if len(calls) > 1 {
		var wg sync.WaitGroup
		for i, call := range calls {
			wg.Add(1)
			go func(idx int, toolUse *types.ToolUse) {
				defer wg.Done()
				a.state.RecordToolCall(toolUse.Name)
				results[idx] = a.dispatcher.Execute(*toolUse)
			}(i, call)
		}
		wg.Wait()
	} else if len(calls) == 1 {
		// Single tool call
		a.state.RecordToolCall(calls[0].Name)
		results[0] = a.dispatcher.Execute(*calls[0])
	}

	return results
}

// StreamQuery processes a user query with streaming response
// Note: This implementation does NOT support tool calls for streaming.
// For tool support, use the non-streaming Query() method instead.
func (a *Agent) StreamQuery(ctx context.Context, input string, outputChan chan<- string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update state
	a.state.IncrementRound()
	a.state.SetProcessing(true)
	defer a.state.SetProcessing(false)

	// Add user message
	if err := a.messages.Add(types.NewTextMessage("user", input)); err != nil {
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// Use non-streaming Query for now since streaming doesn't support tools
	// Create LLM request
	req := a.buildMessageRequest()

	// Send to LLM
	resp, err := a.client.CreateMessage(ctx, req)
	if err != nil {
		a.state.RecordError(err)
		return fmt.Errorf("LLM request failed: %w", err)
	}

	// Add assistant message to history
	if err := a.messages.Add(resp.Message); err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// Support multiple rounds of tool calls
	maxRounds := 10
	currentRound := 0

	for resp.Message.HasToolUse() && currentRound < maxRounds {
		currentRound++
		toolUses := resp.Message.GetToolUses()

		// Execute tools
		results := a.ProcessToolCalls(toolUses)

		// Add tool results to messages
		for _, result := range results {
			if err := a.messages.Add(types.NewToolResultMessage(&result)); err != nil {
				return fmt.Errorf("failed to add tool result: %w", err)
			}
		}

		// Get next response from LLM
		req = a.buildMessageRequest()
		resp, err = a.client.CreateMessage(ctx, req)
		if err != nil {
			a.state.RecordError(err)
			return fmt.Errorf("LLM request failed in round %d: %w", currentRound, err)
		}

		// Add assistant message
		if err := a.messages.Add(resp.Message); err != nil {
			return fmt.Errorf("failed to add assistant message: %w", err)
		}
	}

	// Send final text response through the output channel
	finalText := resp.Message.GetText()
	if finalText != "" {
		outputChan <- finalText
	} else if currentRound > 0 {
		// If tools were executed but no final text, provide feedback
		outputChan <- "✓ Task completed successfully."
	}

	return nil
}

// Reset clears the agent state and message history
func (a *Agent) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.messages.ClearExceptSystem()
	a.state.Reset()
}

// GetStats returns agent statistics
func (a *Agent) GetStats() Stats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return Stats{
		MessageCount:  a.messages.Count(),
		TokenCount:    a.messages.GetTokenCount(),
		ToolCallCount: a.state.GetToolCallCount(),
		ErrorCount:    a.state.GetErrorCount(),
		TotalRounds:   a.state.GetRoundCount(),
		Uptime:        a.state.GetUptime(),
		MessageSummary: a.messages.Summary(),
	}
}

// GetConfig returns the agent configuration
func (a *Agent) GetConfig() *config.Config {
	return a.config
}

// GetMessages returns the message manager
func (a *Agent) GetMessages() *message.Manager {
	return a.messages
}

// GetDispatcher returns the tool dispatcher
func (a *Agent) GetDispatcher() *dispatcher.Dispatcher {
	return a.dispatcher
}

// buildMessageRequest builds an LLM message request
func (a *Agent) buildMessageRequest() llm.MessageRequest {
	// Get tool definitions
	tools := a.getToolDefinitions()

	// Use configured max tokens with validation
	maxTokens := a.config.Model.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096 // Default fallback
	}

	return llm.MessageRequest{
		Model:        a.config.Model.Name,
		Messages:     a.messages.GetHistory(),
		MaxTokens:    maxTokens,
		Temperature:  0.7,
		Stream:       false,
		Tools:        tools,
		SystemPrompt: a.context.GetSystemPrompt(),
	}
}

// getToolDefinitions returns tool definitions for the LLM
func (a *Agent) getToolDefinitions() []llm.ToolDefinition {
	tools := a.dispatcher.ListTools()
	definitions := make([]llm.ToolDefinition, 0, len(tools))

	for _, tool := range tools {
		definitions = append(definitions, llm.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}

	return definitions
}

// initializeSystemPrompt sets up the initial system prompt using PromptManager
func (a *Agent) initializeSystemPrompt() error {
	// Load prompt from various sources (files, env, config, default)
	if err := a.promptManager.Load(); err != nil {
		return fmt.Errorf("failed to load prompt: %w", err)
	}

	// Compose the final system prompt with template variables
	systemPrompt, err := a.promptManager.Compose()
	if err != nil {
		return fmt.Errorf("failed to compose prompt: %w", err)
	}

	// Add system message to message history
	a.messages.AddSystemMessage(systemPrompt)

	return nil
}

// GetPromptManager returns the prompt manager for external access.
func (a *Agent) GetPromptManager() *prompt.Manager {
	return a.promptManager
}

// SetToolObserver registers a tool observer for receiving tool execution events
// The observer will be notified of tool start, completion, and failure events via middleware
// This method is safe to call after agent creation and does not affect existing behavior
// if no observer is set
func (a *Agent) SetToolObserver(obs dispatcher.ToolObserver, opts dispatcher.EventsOptions) {
	if obs == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Add events middleware to dispatcher
	a.dispatcher.AddMiddleware(dispatcher.EventsMiddleware(obs, opts))
}

// Stats represents agent statistics
type Stats struct {
	MessageCount   int
	TokenCount     int
	ToolCallCount  int
	ErrorCount     int
	TotalRounds    int
	Uptime         time.Duration
	MessageSummary string
}

// Builder provides a fluent interface for building an Agent
type Builder struct {
	config *config.Config
	tools  []interface{}
}

// NewBuilder creates a new agent builder
func NewBuilder() *Builder {
	return &Builder{
		config: config.DefaultConfig(),
		tools:  []interface{}{},
	}
}

// WithConfig sets the configuration
func (b *Builder) WithConfig(cfg *config.Config) *Builder {
	b.config = cfg
	return b
}

// WithLLM sets the LLM provider
func (b *Builder) WithLLM(provider, apiKey, model string) *Builder {
	b.config.Model.Provider = provider
	b.config.Model.APIKey = apiKey
	b.config.Model.Name = model
	return b
}

// WithWorkDir sets the working directory
func (b *Builder) WithWorkDir(dir string) *Builder {
	b.config.WorkDir = dir
	return b
}

// WithTools adds tools to the agent
func (b *Builder) WithTools(tools ...interface{}) *Builder {
	b.tools = append(b.tools, tools...)
	return b
}

// Build creates the agent
func (b *Builder) Build() (*Agent, error) {
	agent, err := NewAgent(b.config)
	if err != nil {
		return nil, err
	}

	// Register tools
	// Note: Tool registration would be done here if tools were provided

	return agent, nil
}