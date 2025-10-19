package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/config"
	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
)

// MockLLMClient is a mock implementation of llm.Client for testing
type MockLLMClient struct {
	responses   []llm.MessageResponse
	responseIdx int
	streamChunks []llm.StreamChunk
	model       string
	available   bool
	error       error
}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		model:     "mock-model",
		available: true,
		responses: []llm.MessageResponse{},
	}
}

func (m *MockLLMClient) CreateMessage(ctx context.Context, req llm.MessageRequest) (*llm.MessageResponse, error) {
	if m.error != nil {
		return nil, m.error
	}

	if m.responseIdx >= len(m.responses) {
		return &llm.MessageResponse{
			ID:        "mock-response-1",
			Model:     m.model,
			Message:   types.NewTextMessage("assistant", "Mock response"),
			CreatedAt: time.Now(),
		}, nil
	}

	resp := m.responses[m.responseIdx]
	m.responseIdx++
	return &resp, nil
}

func (m *MockLLMClient) StreamMessage(ctx context.Context, req llm.MessageRequest) (<-chan llm.StreamChunk, error) {
	if m.error != nil {
		return nil, m.error
	}

	ch := make(chan llm.StreamChunk)
	go func() {
		defer close(ch)
		for _, chunk := range m.streamChunks {
			select {
			case <-ctx.Done():
				return
			case ch <- chunk:
			}
		}
	}()
	return ch, nil
}

func (m *MockLLMClient) GetModel() string {
	return m.model
}

func (m *MockLLMClient) SetModel(model string) error {
	m.model = model
	return nil
}

func (m *MockLLMClient) IsAvailable() bool {
	return m.available
}

// Test helper to create a test config
func createTestConfig(t *testing.T) *config.Config {
	tempDir := t.TempDir()

	return &config.Config{
		Model: config.ModelConfig{
			Provider:  "mock",
			APIKey:    "test-key",
			Name:      "test-model",
			MaxTokens: 1000,
			Timeout:   10,
		},
		WorkDir: tempDir,
		Tools: config.ToolsConfig{
			Enabled: []string{"bash", "file_read", "file_write"},
		},
		Todo: config.TodoConfig{
			MaxItems:         20,
			ReminderInterval: 5,
		},
		Output: config.OutputConfig{
			MaxChars:    10000,
			Format:      "markdown",
			Colors:      true,
			ShowSpinner: false,
		},
	}
}

func TestNewAgent(t *testing.T) {
	// Register mock client factory
	llm.RegisterClientFactory("mock", func(config llm.ClientConfig) (llm.Client, error) {
		return NewMockLLMClient(), nil
	})

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  createTestConfig(t),
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && agent == nil {
				t.Error("NewAgent() returned nil agent without error")
			}
		})
	}
}

func TestAgent_Query(t *testing.T) {
	// Register mock client factory
	llm.RegisterClientFactory("mock", func(config llm.ClientConfig) (llm.Client, error) {
		client := NewMockLLMClient()
		client.responses = []llm.MessageResponse{
			{
				ID:        "test-1",
				Model:     "mock-model",
				Message:   types.NewTextMessage("assistant", "Test response"),
				CreatedAt: time.Now(),
			},
		}
		return client, nil
	})

	cfg := createTestConfig(t)
	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()
	response, err := agent.Query(ctx, "Test query")
	if err != nil {
		t.Errorf("Query() error = %v", err)
	}

	if response != "Test response" {
		t.Errorf("Query() = %v, want %v", response, "Test response")
	}
}

func TestAgent_Reset(t *testing.T) {
	// Register mock client factory
	llm.RegisterClientFactory("mock", func(config llm.ClientConfig) (llm.Client, error) {
		return NewMockLLMClient(), nil
	})

	cfg := createTestConfig(t)
	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add some messages
	agent.messages.AddUserMessage("Test message 1")
	agent.messages.AddAssistantMessage("Response 1")

	// Check message count before reset
	if agent.messages.Count() < 2 {
		t.Error("Expected at least 2 messages before reset")
	}

	// Reset
	agent.Reset()

	// Check that non-system messages were cleared
	count := agent.messages.Count()
	for i := 0; i < count; i++ {
		msg := agent.messages.GetHistory()[i]
		if msg.Role != "system" {
			t.Error("Reset() should clear all non-system messages")
		}
	}
}

func TestAgent_GetStats(t *testing.T) {
	// Register mock client factory
	llm.RegisterClientFactory("mock", func(config llm.ClientConfig) (llm.Client, error) {
		return NewMockLLMClient(), nil
	})

	cfg := createTestConfig(t)
	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add some activity
	agent.messages.AddUserMessage("Test message")
	agent.state.IncrementRound()
	agent.state.RecordToolCall("test_tool")

	stats := agent.GetStats()

	if stats.MessageCount == 0 {
		t.Error("GetStats() MessageCount should be > 0")
	}

	if stats.TotalRounds != 1 {
		t.Errorf("GetStats() TotalRounds = %v, want 1", stats.TotalRounds)
	}

	if stats.ToolCallCount != 1 {
		t.Errorf("GetStats() ToolCallCount = %v, want 1", stats.ToolCallCount)
	}
}

func TestState_RoundCounter(t *testing.T) {
	state := NewState()

	// Initial count should be 0
	if state.GetRoundCount() != 0 {
		t.Errorf("Initial round count = %v, want 0", state.GetRoundCount())
	}

	// Increment and check
	for i := 1; i <= 5; i++ {
		state.IncrementRound()
		if state.GetRoundCount() != i {
			t.Errorf("After %d increments, count = %v, want %v", i, state.GetRoundCount(), i)
		}
	}
}

func TestState_ToolCallTracking(t *testing.T) {
	state := NewState()

	// Record tool calls
	state.RecordToolCall("bash")
	state.RecordToolCall("file_read")
	state.RecordToolCall("bash")

	// Check total count
	if state.GetToolCallCount() != 3 {
		t.Errorf("Tool call count = %v, want 3", state.GetToolCallCount())
	}

	// Check individual tool stats
	stats := state.GetToolCallStats()
	if stats["bash"] != 2 {
		t.Errorf("Bash tool count = %v, want 2", stats["bash"])
	}
	if stats["file_read"] != 1 {
		t.Errorf("File_read tool count = %v, want 1", stats["file_read"])
	}
}

func TestState_ErrorTracking(t *testing.T) {
	state := NewState()

	// Initially no errors
	if state.HasErrors() {
		t.Error("Initial state should have no errors")
	}

	// Record some errors
	err1 := fmt.Errorf("test error 1")
	err2 := fmt.Errorf("test error 2")

	state.RecordError(err1)
	state.RecordError(err2)

	// Check error count
	if state.GetErrorCount() != 2 {
		t.Errorf("Error count = %v, want 2", state.GetErrorCount())
	}

	// Check last error
	if state.GetLastError() != err2 {
		t.Error("GetLastError() should return the most recent error")
	}

	// Check error log
	errorLog := state.GetErrorLog()
	if len(errorLog) != 2 {
		t.Errorf("Error log length = %v, want 2", len(errorLog))
	}
}

func TestState_RecoveryPoints(t *testing.T) {
	state := NewState()

	// Create recovery points
	data1 := map[string]interface{}{"round": 1}
	id1 := state.CreateRecoveryPoint(data1)

	data2 := map[string]interface{}{"round": 2}
	id2 := state.CreateRecoveryPoint(data2)

	// Get recovery point by ID
	point := state.GetRecoveryPoint(id1)
	if point == nil {
		t.Error("Should be able to retrieve recovery point by ID")
	}

	// Get latest recovery point
	latest := state.GetLatestRecoveryPoint()
	if latest == nil || latest.ID != id2 {
		t.Error("GetLatestRecoveryPoint() should return the most recent point")
	}
}

func TestContext_Creation(t *testing.T) {
	tempDir := t.TempDir()
	ctx := NewContext(tempDir)

	if ctx.GetWorkDir() != tempDir {
		t.Errorf("Work directory = %v, want %v", ctx.GetWorkDir(), tempDir)
	}

	if ctx.GetSystemPrompt() == "" {
		t.Error("System prompt should not be empty")
	}

	projectInfo := ctx.GetProjectInfo()
	if projectInfo == nil {
		t.Error("Project info should not be nil")
	}
}

func TestContext_ProjectAnalysis(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files
	goFile := filepath.Join(tempDir, "test.go")
	pyFile := filepath.Join(tempDir, "test.py")

	os.WriteFile(goFile, []byte("package main"), 0644)
	os.WriteFile(pyFile, []byte("print('hello')"), 0644)

	ctx := NewContext(tempDir)
	projectInfo := ctx.GetProjectInfo()

	if projectInfo.FileCount != 2 {
		t.Errorf("File count = %v, want 2", projectInfo.FileCount)
	}

	// Language detection depends on which has more files
	// In this case, both have 1 file each
	validLanguages := []string{"Go", "Python", "Unknown"}
	found := false
	for _, lang := range validLanguages {
		if projectInfo.Language == lang {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Unexpected language detected: %v", projectInfo.Language)
	}
}

func TestBuilder(t *testing.T) {
	// Register mock client factory
	llm.RegisterClientFactory("mock", func(config llm.ClientConfig) (llm.Client, error) {
		return NewMockLLMClient(), nil
	})

	tempDir := t.TempDir()

	agent, err := NewBuilder().
		WithLLM("mock", "test-key", "test-model").
		WithWorkDir(tempDir).
		Build()

	if err != nil {
		t.Fatalf("Builder.Build() error = %v", err)
	}

	if agent == nil {
		t.Error("Builder.Build() returned nil agent")
	}

	if agent.config.Model.Provider != "mock" {
		t.Errorf("Provider = %v, want mock", agent.config.Model.Provider)
	}

	if agent.config.WorkDir != tempDir {
		t.Errorf("WorkDir = %v, want %v", agent.config.WorkDir, tempDir)
	}
}