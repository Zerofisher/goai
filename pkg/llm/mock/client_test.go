package mock

import (
	"context"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
)

func TestNewClient(t *testing.T) {
	responses := []*llm.MessageResponse{
		{
			ID:      "test-1",
			Model:   "mock-model",
			Message: types.NewTextMessage("assistant", "Hello"),
		},
	}

	client := NewClient(responses, nil)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.GetModel() != "mock-model" {
		t.Errorf("Expected model 'mock-model', got '%s'", client.GetModel())
	}

	if client.Provider() != "mock" {
		t.Errorf("Expected provider 'mock', got '%s'", client.Provider())
	}

	if !client.IsAvailable() {
		t.Error("Expected client to be available")
	}
}

func TestNewSimpleClient(t *testing.T) {
	client := NewSimpleClient("Test response")

	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "Hello"),
		},
	}

	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp.Message.GetText() != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", resp.Message.GetText())
	}
}

func TestClient_CreateMessage(t *testing.T) {
	responses := []*llm.MessageResponse{
		{
			ID:      "test-1",
			Model:   "mock-model",
			Message: types.NewTextMessage("assistant", "Response 1"),
		},
		{
			ID:      "test-2",
			Model:   "mock-model",
			Message: types.NewTextMessage("assistant", "Response 2"),
		},
	}

	client := NewClient(responses, nil)
	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "Hello"),
		},
	}

	// First response
	resp1, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp1.Message.GetText() != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", resp1.Message.GetText())
	}

	// Second response
	resp2, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp2.Message.GetText() != "Response 2" {
		t.Errorf("Expected 'Response 2', got '%s'", resp2.Message.GetText())
	}

	// No more responses
	_, err = client.CreateMessage(ctx, req)
	if err == nil {
		t.Error("Expected error when no more responses available")
	}
}

func TestClient_StreamMessage(t *testing.T) {
	chunks := []llm.StreamChunk{
		{
			ID:    "chunk-1",
			Model: "mock-model",
			Delta: types.Content{Type: "text", Text: "Hello"},
			Done:  false,
		},
		{
			ID:    "chunk-2",
			Model: "mock-model",
			Delta: types.Content{Type: "text", Text: " World"},
			Done:  false,
		},
	}

	client := NewClient(nil, chunks)
	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "Hello"),
		},
	}

	chunkChan, err := client.StreamMessage(ctx, req)
	if err != nil {
		t.Fatalf("StreamMessage failed: %v", err)
	}

	receivedChunks := 0
	fullText := ""

	for chunk := range chunkChan {
		if chunk.Error != nil {
			t.Fatalf("Received error chunk: %v", chunk.Error)
		}

		if chunk.Done {
			break
		}

		fullText += chunk.Delta.Text
		receivedChunks++
	}

	if receivedChunks != 2 {
		t.Errorf("Expected 2 chunks, got %d", receivedChunks)
	}

	if fullText != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", fullText)
	}
}

func TestClient_CountTokens(t *testing.T) {
	client := NewClient(nil, nil)
	ctx := context.Background()

	req := llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "This is a test message"),
		},
	}

	count, err := client.CountTokens(ctx, req)
	if err != nil {
		t.Fatalf("CountTokens failed: %v", err)
	}

	// Simple mock divides character count by 4
	expectedCount := len("This is a test message") / 4
	if count != expectedCount {
		t.Errorf("Expected %d tokens, got %d", expectedCount, count)
	}
}

func TestClient_CreateBatch(t *testing.T) {
	responses := []*llm.MessageResponse{
		{
			ID:      "batch-1",
			Model:   "mock-model",
			Message: types.NewTextMessage("assistant", "Batch response"),
		},
	}

	client := NewClient(responses, nil)
	ctx := context.Background()

	requests := []llm.MessageRequest{
		{
			Messages: []types.Message{
				types.NewTextMessage("user", "Request 1"),
			},
		},
		{
			Messages: []types.Message{
				types.NewTextMessage("user", "Request 2"),
			},
		},
	}

	batch, err := client.CreateBatch(ctx, requests)
	if err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	if batch.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", batch.Status)
	}

	if batch.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", batch.TotalRequests)
	}

	if batch.CompletedCount != 2 {
		t.Errorf("Expected 2 completed, got %d", batch.CompletedCount)
	}
}

func TestClient_GetBatch(t *testing.T) {
	client := NewClient(nil, nil)
	ctx := context.Background()

	batch, err := client.GetBatch(ctx, "test-batch-id")
	if err != nil {
		t.Fatalf("GetBatch failed: %v", err)
	}

	if batch.ID != "test-batch-id" {
		t.Errorf("Expected ID 'test-batch-id', got '%s'", batch.ID)
	}

	if batch.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", batch.Status)
	}

	if batch.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestClient_CustomFunctions(t *testing.T) {
	client := NewClient(nil, nil)
	ctx := context.Background()

	// Test custom CreateMessage function
	customResponseCalled := false
	client.CreateMessageFunc = func(ctx context.Context, req llm.MessageRequest) (*llm.MessageResponse, error) {
		customResponseCalled = true
		return &llm.MessageResponse{
			ID:        "custom-1",
			Model:     "custom-model",
			Message:   types.NewTextMessage("assistant", "Custom response"),
			CreatedAt: time.Now(),
		}, nil
	}

	resp, err := client.CreateMessage(ctx, llm.MessageRequest{})
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if !customResponseCalled {
		t.Error("Custom function was not called")
	}

	if resp.Message.GetText() != "Custom response" {
		t.Errorf("Expected 'Custom response', got '%s'", resp.Message.GetText())
	}
}

func TestClient_AddResponse(t *testing.T) {
	client := NewClient(nil, nil)

	newResp := &llm.MessageResponse{
		ID:      "added-1",
		Model:   "mock-model",
		Message: types.NewTextMessage("assistant", "Added response"),
	}

	client.AddResponse(newResp)

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, llm.MessageRequest{})
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp.ID != "added-1" {
		t.Errorf("Expected ID 'added-1', got '%s'", resp.ID)
	}
}

func TestClient_Reset(t *testing.T) {
	responses := []*llm.MessageResponse{
		{
			ID:      "test-1",
			Model:   "mock-model",
			Message: types.NewTextMessage("assistant", "Response 1"),
		},
	}

	client := NewClient(responses, nil)
	ctx := context.Background()
	req := llm.MessageRequest{}

	// Consume the first response
	_, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	// Reset and consume again
	client.Reset()
	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage after Reset failed: %v", err)
	}

	if resp.ID != "test-1" {
		t.Errorf("Expected ID 'test-1' after reset, got '%s'", resp.ID)
	}
}

func TestClient_SetModel(t *testing.T) {
	client := NewClient(nil, nil)

	err := client.SetModel("new-model")
	if err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	if client.GetModel() != "new-model" {
		t.Errorf("Expected model 'new-model', got '%s'", client.GetModel())
	}
}

func TestClient_Close(t *testing.T) {
	client := NewClient(nil, nil)

	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
