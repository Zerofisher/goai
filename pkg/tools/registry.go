package tools

import (
	"fmt"
	"sync"
)

// DefaultToolRegistry is the default implementation of ToolRegistry
type DefaultToolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *DefaultToolRegistry {
	return &DefaultToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *DefaultToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}
	
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *DefaultToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tools
func (r *DefaultToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	
	return tools
}

// ListByCategory returns tools filtered by category
func (r *DefaultToolRegistry) ListByCategory(category string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var tools []Tool
	for _, tool := range r.tools {
		if tool.Category() == category {
			tools = append(tools, tool)
		}
	}
	
	return tools
}