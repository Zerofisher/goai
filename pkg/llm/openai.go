package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

func init() {
	// Register OpenAI factory
	RegisterClientFactory("openai", NewOpenAIClient)
}

// OpenAIClient implements the Client interface for OpenAI API
type OpenAIClient struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config ClientConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, types.NewAgentError(types.ErrCodeLLMInvalidKey, "OpenAI API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-4"
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// CreateMessage sends a message to OpenAI and returns the response
func (c *OpenAIClient) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Convert our message format to OpenAI format
	openAIReq := c.buildOpenAIRequest(req)

	// Marshal request
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMParsing, "failed to marshal request", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMConnection, "failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMConnection, "failed to send request", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, types.NewAgentError(types.ErrCodeLLMConnection, fmt.Sprintf("OpenAI API error: %d - %s", resp.StatusCode, string(body)))
	}

	// Parse response
	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, types.WrapError(types.ErrCodeLLMParsing, "failed to parse response", err)
	}

	// Convert to our format
	return c.convertResponse(&openAIResp), nil
}

// StreamMessage sends a message to OpenAI and streams the response
func (c *OpenAIClient) StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error) {
	// Set stream to true
	req.Stream = true
	openAIReq := c.buildOpenAIRequest(req)
	openAIReq["stream"] = true

	// Marshal request
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMParsing, "failed to marshal request", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMConnection, "failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.WrapError(types.ErrCodeLLMConnection, "failed to send request", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, types.NewAgentError(types.ErrCodeLLMConnection, fmt.Sprintf("OpenAI API error: %d - %s", resp.StatusCode, string(body)))
	}

	// Create channel for streaming
	ch := make(chan StreamChunk, 100)

	// Start goroutine to read stream
	go c.handleStream(resp.Body, ch)

	return ch, nil
}

// GetModel returns the current model being used
func (c *OpenAIClient) GetModel() string {
	return c.config.Model
}

// SetModel sets the model to use
func (c *OpenAIClient) SetModel(model string) error {
	c.config.Model = model
	return nil
}

// IsAvailable checks if the client is available and configured
func (c *OpenAIClient) IsAvailable() bool {
	return c.config.APIKey != ""
}

// buildOpenAIRequest converts our request format to OpenAI format
func (c *OpenAIClient) buildOpenAIRequest(req MessageRequest) map[string]interface{} {
	openAIReq := map[string]interface{}{
		"model": c.config.Model,
	}

	if req.MaxTokens > 0 {
		openAIReq["max_tokens"] = req.MaxTokens
	} else if c.config.MaxTokens > 0 {
		openAIReq["max_tokens"] = c.config.MaxTokens
	}

	if req.Temperature > 0 {
		openAIReq["temperature"] = req.Temperature
	} else {
		openAIReq["temperature"] = c.config.Temperature
	}

	// Convert messages
	openAIMessages := c.convertMessages(req.Messages, req.SystemPrompt)
	openAIReq["messages"] = openAIMessages

	// Add tools if any
	if len(req.Tools) > 0 {
		openAITools := c.convertTools(req.Tools)
		openAIReq["tools"] = openAITools
		openAIReq["tool_choice"] = "auto"
	}

	return openAIReq
}

// convertMessages converts our message format to OpenAI format
func (c *OpenAIClient) convertMessages(messages []types.Message, systemPrompt string) []map[string]interface{} {
	openAIMessages := make([]map[string]interface{}, 0)

	// Add system prompt if provided
	if systemPrompt != "" {
		openAIMessages = append(openAIMessages, map[string]interface{}{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	for _, msg := range messages {
		openAIMsg := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			// Simple text message
			openAIMsg["content"] = msg.Content[0].Text
		} else {
			// Complex content
			content := make([]map[string]interface{}, 0)
			for _, c := range msg.Content {
				switch c.Type {
				case "text":
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": c.Text,
					})
				case "tool_use":
					if c.ToolUse != nil {
						// Convert arguments to JSON string as required by OpenAI API
						argsJSON, err := json.Marshal(c.ToolUse.Input)
						if err != nil {
							// Fallback to empty object if marshal fails
							argsJSON = []byte("{}")
						}

						// Convert to OpenAI tool call format
						openAIMsg["tool_calls"] = []map[string]interface{}{
							{
								"id":   c.ToolUse.ID,
								"type": "function",
								"function": map[string]interface{}{
									"name":      c.ToolUse.Name,
									"arguments": string(argsJSON),
								},
							},
						}
					}
				case "tool_result":
					if c.ToolResult != nil {
						// Convert to OpenAI tool response format
						openAIMsg["role"] = "tool"
						openAIMsg["tool_call_id"] = c.ToolResult.ToolUseID
						openAIMsg["content"] = c.ToolResult.Content
					}
				}
			}
			if len(content) > 0 {
				openAIMsg["content"] = content
			}
		}

		openAIMessages = append(openAIMessages, openAIMsg)
	}

	return openAIMessages
}

// convertTools converts our tool format to OpenAI format
func (c *OpenAIClient) convertTools(tools []ToolDefinition) []map[string]interface{} {
	openAITools := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		openAITools = append(openAITools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.InputSchema,
			},
		})
	}
	return openAITools
}

// convertResponse converts OpenAI response to our format
func (c *OpenAIClient) convertResponse(resp *openAIResponse) *MessageResponse {
	if len(resp.Choices) == 0 {
		return nil
	}

	choice := resp.Choices[0]
	msg := types.Message{
		Role:    choice.Message.Role,
		Content: []types.Content{},
	}

	// Handle content
	if choice.Message.Content != "" {
		msg.Content = append(msg.Content, types.Content{
			Type: "text",
			Text: choice.Message.Content,
		})
	}

	// Handle tool calls
	for _, toolCall := range choice.Message.ToolCalls {
		// Parse arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			args = map[string]interface{}{"raw": toolCall.Function.Arguments}
		}

		msg.Content = append(msg.Content, types.Content{
			Type: "tool_use",
			ToolUse: &types.ToolUse{
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: args,
			},
		})
	}

	return &MessageResponse{
		ID:        resp.ID,
		Model:     resp.Model,
		Message:   msg,
		Usage:     c.convertUsage(&resp.Usage),
		CreatedAt: time.Unix(resp.Created, 0),
	}
}

// convertUsage converts OpenAI usage to our format
func (c *OpenAIClient) convertUsage(usage *openAIUsage) *TokenUsage {
	if usage == nil {
		return nil
	}
	return &TokenUsage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

// handleStream handles streaming responses from OpenAI
func (c *OpenAIClient) handleStream(body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer func() { _ = body.Close() }()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- StreamChunk{Done: true}
			return
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamChunk{Error: err}
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			streamChunk := StreamChunk{
				ID:    chunk.ID,
				Model: chunk.Model,
			}

			if delta.Content != "" {
				streamChunk.Delta = types.Content{
					Type: "text",
					Text: delta.Content,
				}
			}

			if len(delta.ToolCalls) > 0 {
				// Handle tool calls in stream
				toolCall := delta.ToolCalls[0]
				var args map[string]interface{}
				if toolCall.Function.Arguments != "" {
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						args = map[string]interface{}{"raw": toolCall.Function.Arguments}
					}
				}

				streamChunk.Delta = types.Content{
					Type: "tool_use",
					ToolUse: &types.ToolUse{
						ID:    toolCall.ID,
						Name:  toolCall.Function.Name,
						Input: args,
					},
				}
			}

			ch <- streamChunk
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Error: err}
	}
}

// OpenAI API response structures

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Index        int            `json:"index"`
	Message      openAIMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function openAIFunctionCall   `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIStreamChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []openAIStreamChoice `json:"choices"`
}

type openAIStreamChoice struct {
	Index        int              `json:"index"`
	Delta        openAIStreamDelta `json:"delta"`
	FinishReason *string          `json:"finish_reason"`
}

type openAIStreamDelta struct {
	Role      string           `json:"role,omitempty"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}