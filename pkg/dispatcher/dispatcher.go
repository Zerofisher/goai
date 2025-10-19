package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zerofisher/goai/pkg/tools"
	"github.com/Zerofisher/goai/pkg/types"
)

// Dispatcher manages and executes tools
type Dispatcher struct {
	registry     tools.Registry
	security     tools.SecurityValidator
	workDir      string
	maxParallel  int
	timeout      time.Duration
	middlewares  []Middleware
	mu           sync.RWMutex
}

// Middleware is a function that wraps tool execution
type Middleware func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult

// ExecuteFunc is the function signature for tool execution
type ExecuteFunc func(ctx context.Context, toolUse types.ToolUse) types.ToolResult

// New creates a new tool dispatcher
func New(workDir string) *Dispatcher {
	return &Dispatcher{
		registry:    tools.NewRegistry(),
		security:    tools.NewSecurityValidator(workDir),
		workDir:     workDir,
		maxParallel: 5,
		timeout:     30 * time.Second,
		middlewares: []Middleware{},
	}
}

// Register registers a tool with the dispatcher
func (d *Dispatcher) Register(tool tools.Tool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.registry.Register(tool)
}

// RegisterAll registers multiple tools
func (d *Dispatcher) RegisterAll(tools ...tools.Tool) error {
	for _, tool := range tools {
		if err := d.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}
	}
	return nil
}

// Execute executes a single tool use
func (d *Dispatcher) Execute(toolUse types.ToolUse) types.ToolResult {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	return d.ExecuteWithContext(ctx, toolUse)
}

// ExecuteWithContext executes a single tool use with context
func (d *Dispatcher) ExecuteWithContext(ctx context.Context, toolUse types.ToolUse) types.ToolResult {
	// Validate tool use
	if err := toolUse.Validate(); err != nil {
		return *toolUse.Error(fmt.Errorf("invalid tool use: %w", err))
	}

	// Apply middlewares
	execute := d.executeCore
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		middleware := d.middlewares[i]
		next := execute
		execute = func(ctx context.Context, tu types.ToolUse) types.ToolResult {
			return middleware(ctx, tu, next)
		}
	}

	return execute(ctx, toolUse)
}

// executeCore is the core execution logic
func (d *Dispatcher) executeCore(ctx context.Context, toolUse types.ToolUse) types.ToolResult {
	// Get the tool
	d.mu.RLock()
	tool, err := d.registry.Get(toolUse.Name)
	d.mu.RUnlock()

	if err != nil {
		return *toolUse.Error(fmt.Errorf("tool not found: %s", toolUse.Name))
	}

	// Security check
	if err := d.security.CheckPermission(toolUse.Name, toolUse.Input); err != nil {
		return *toolUse.Error(fmt.Errorf("security check failed: %w", err))
	}

	// Validate input
	if err := tool.Validate(toolUse.Input); err != nil {
		return *toolUse.Error(fmt.Errorf("input validation failed: %w", err))
	}

	// Execute the tool
	result, err := tool.Execute(ctx, toolUse.Input)
	if err != nil {
		return *toolUse.Error(err)
	}

	return *toolUse.Success(result)
}

// ExecuteBatch executes multiple tool uses
func (d *Dispatcher) ExecuteBatch(toolUses []types.ToolUse) []types.ToolResult {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout*time.Duration(len(toolUses)))
	defer cancel()

	return d.ExecuteBatchWithContext(ctx, toolUses)
}

// ExecuteBatchWithContext executes multiple tool uses with context
func (d *Dispatcher) ExecuteBatchWithContext(ctx context.Context, toolUses []types.ToolUse) []types.ToolResult {
	if len(toolUses) == 0 {
		return []types.ToolResult{}
	}

	// Execute sequentially if only one or parallel limit is 1
	if len(toolUses) == 1 || d.maxParallel <= 1 {
		results := make([]types.ToolResult, 0, len(toolUses))
		for _, toolUse := range toolUses {
			results = append(results, d.ExecuteWithContext(ctx, toolUse))
		}
		return results
	}

	// Execute in parallel with limit
	results := make([]types.ToolResult, len(toolUses))
	semaphore := make(chan struct{}, d.maxParallel)
	var wg sync.WaitGroup

	for i, toolUse := range toolUses {
		wg.Add(1)
		go func(idx int, tu types.ToolUse) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check context cancellation
			select {
			case <-ctx.Done():
				results[idx] = *tu.Error(fmt.Errorf("execution cancelled: %w", ctx.Err()))
				return
			default:
			}

			// Execute tool
			results[idx] = d.ExecuteWithContext(ctx, tu)
		}(i, toolUse)
	}

	wg.Wait()
	return results
}

// SetMaxParallel sets the maximum number of parallel tool executions
func (d *Dispatcher) SetMaxParallel(max int) {
	if max <= 0 {
		max = 1
	}
	d.mu.Lock()
	d.maxParallel = max
	d.mu.Unlock()
}

// SetTimeout sets the timeout for tool execution
func (d *Dispatcher) SetTimeout(timeout time.Duration) {
	d.mu.Lock()
	d.timeout = timeout
	d.mu.Unlock()
}

// AddMiddleware adds a middleware to the dispatcher
func (d *Dispatcher) AddMiddleware(middleware Middleware) {
	d.mu.Lock()
	d.middlewares = append(d.middlewares, middleware)
	d.mu.Unlock()
}

// GetRegistry returns the tool registry
func (d *Dispatcher) GetRegistry() tools.Registry {
	return d.registry
}

// GetSecurity returns the security validator
func (d *Dispatcher) GetSecurity() tools.SecurityValidator {
	return d.security
}

// ListTools returns a list of all registered tools
func (d *Dispatcher) ListTools() []tools.Tool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.registry.List()
}

// HasTool checks if a tool is registered
func (d *Dispatcher) HasTool(name string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.registry.Has(name)
}

// Clear removes all registered tools
func (d *Dispatcher) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.registry.Clear()
}

// Stats returns execution statistics
type Stats struct {
	RegisteredTools int
	MaxParallel     int
	Timeout         time.Duration
	MiddlewareCount int
}

// GetStats returns dispatcher statistics
func (d *Dispatcher) GetStats() Stats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return Stats{
		RegisteredTools: len(d.registry.List()),
		MaxParallel:     d.maxParallel,
		Timeout:         d.timeout,
		MiddlewareCount: len(d.middlewares),
	}
}