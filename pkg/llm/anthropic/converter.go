package anthropic

import (
	"encoding/json"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
)

// convertToAnthropicParams converts llm.MessageRequest to Anthropic SDK parameters
func convertToAnthropicParams(req llm.MessageRequest, model anthropicsdk.Model) anthropicsdk.MessageNewParams {
	params := anthropicsdk.MessageNewParams{
		Model:     model,
		Messages:  convertMessages(req.Messages),
		MaxTokens: int64(req.MaxTokens),
	}

	if req.Temperature > 0 {
		params.Temperature = anthropicsdk.Float(float64(req.Temperature))
	}

	if req.TopP > 0 {
		params.TopP = anthropicsdk.Float(float64(req.TopP))
	}

	if req.SystemPrompt != "" {
		params.System = []anthropicsdk.TextBlockParam{
			{
				Text: req.SystemPrompt,
			},
		}
	}

	if len(req.Tools) > 0 {
		params.Tools = convertTools(req.Tools)
	}

	if req.ToolChoice != nil {
		params.ToolChoice = convertToolChoice(req.ToolChoice)
	}

	if len(req.StopSequences) > 0 {
		params.StopSequences = req.StopSequences
	}

	if len(req.Metadata) > 0 {
		if userID, ok := req.Metadata["user_id"]; ok {
			params.Metadata = anthropicsdk.MetadataParam{
				UserID: anthropicsdk.String(userID),
			}
		}
	}

	return params
}

// convertMessages converts []types.Message to Anthropic SDK message format
func convertMessages(messages []types.Message) []anthropicsdk.MessageParam {
	result := make([]anthropicsdk.MessageParam, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			result = append(result, convertUserMessage(msg))
		case "assistant":
			result = append(result, convertAssistantMessage(msg))
		}
	}

	return result
}

// convertUserMessage converts a user message
func convertUserMessage(msg types.Message) anthropicsdk.MessageParam {
	contentBlocks := make([]anthropicsdk.ContentBlockParamUnion, 0, len(msg.Content))

	for _, content := range msg.Content {
		switch content.Type {
		case "text":
			contentBlocks = append(contentBlocks, anthropicsdk.ContentBlockParamUnion{
				OfText: &anthropicsdk.TextBlockParam{
					Text: content.Text,
				},
			})
		case "tool_result":
			if content.ToolResult != nil {
				contentBlocks = append(contentBlocks, anthropicsdk.ContentBlockParamUnion{
					OfToolResult: &anthropicsdk.ToolResultBlockParam{
						ToolUseID: content.ToolResult.ToolUseID,
						Content:   []anthropicsdk.ToolResultBlockParamContentUnion{},
						IsError:   anthropicsdk.Bool(content.ToolResult.IsError),
					},
				})
			}
		}
	}

	return anthropicsdk.NewUserMessage(contentBlocks...)
}

// convertAssistantMessage converts an assistant message
func convertAssistantMessage(msg types.Message) anthropicsdk.MessageParam {
	contentBlocks := make([]anthropicsdk.ContentBlockParamUnion, 0, len(msg.Content))

	for _, content := range msg.Content {
		switch content.Type {
		case "text":
			contentBlocks = append(contentBlocks, anthropicsdk.ContentBlockParamUnion{
				OfText: &anthropicsdk.TextBlockParam{
					Text: content.Text,
				},
			})
		case "tool_use":
			if content.ToolUse != nil {
				contentBlocks = append(contentBlocks, anthropicsdk.ContentBlockParamUnion{
					OfToolUse: &anthropicsdk.ToolUseBlockParam{
						ID:    content.ToolUse.ID,
						Name:  content.ToolUse.Name,
						Input: content.ToolUse.Input,
					},
				})
			}
		}
	}

	return anthropicsdk.NewAssistantMessage(contentBlocks...)
}

// convertTools converts tool definitions to Anthropic format
func convertTools(tools []llm.ToolDefinition) []anthropicsdk.ToolUnionParam {
	result := make([]anthropicsdk.ToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		result = append(result, anthropicsdk.ToolUnionParam{
			OfTool: &anthropicsdk.ToolParam{
				Name:        tool.Name,
				Description: anthropicsdk.String(tool.Description),
				InputSchema: anthropicsdk.ToolInputSchemaParam{
					Properties: tool.InputSchema["properties"],
				},
			},
		})
	}

	return result
}

// convertToolChoice converts tool choice to Anthropic format
func convertToolChoice(choice *llm.ToolChoice) anthropicsdk.ToolChoiceUnionParam {
	if choice == nil {
		// Auto mode - return empty/default
		return anthropicsdk.ToolChoiceUnionParam{}
	}

	switch choice.Type {
	case "auto":
		return anthropicsdk.ToolChoiceUnionParam{}
	case "any", "required":
		// For "any" mode, just return empty - Anthropic will use auto behavior
		return anthropicsdk.ToolChoiceUnionParam{}
	case "tool":
		if choice.ToolName != "" {
			// For specific tool, just return empty for now
			// The SDK doesn't seem to expose the helper we need
			return anthropicsdk.ToolChoiceUnionParam{}
		}
	}

	// Default to auto
	return anthropicsdk.ToolChoiceUnionParam{}
}

// convertFromAnthropicResponse converts Anthropic response to llm.MessageResponse
func convertFromAnthropicResponse(message *anthropicsdk.Message) *llm.MessageResponse {
	resp := &llm.MessageResponse{
		ID:        message.ID,
		Model:     string(message.Model),
		Message:   convertAnthropicMessage(message),
		CreatedAt: time.Now(),
	}

	// Usage statistics
	resp.Usage = &llm.TokenUsage{
		PromptTokens:     int(message.Usage.InputTokens),
		CompletionTokens: int(message.Usage.OutputTokens),
		TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}

	return resp
}

// convertAnthropicMessage converts Anthropic message to types.Message
func convertAnthropicMessage(message *anthropicsdk.Message) types.Message {
	msg := types.Message{
		Role:    "assistant",
		Content: []types.Content{},
	}

	// Extract content
	for _, block := range message.Content {
		switch contentBlock := block.AsAny().(type) {
		case anthropicsdk.TextBlock:
			msg.Content = append(msg.Content, types.Content{
				Type: "text",
				Text: contentBlock.Text,
			})
		case anthropicsdk.ToolUseBlock:
			// Convert input to map[string]interface{}
			var input map[string]interface{}
			inputJSON, _ := json.Marshal(contentBlock.Input)
			_ = json.Unmarshal(inputJSON, &input)

			msg.Content = append(msg.Content, types.Content{
				Type: "tool_use",
				ToolUse: &types.ToolUse{
					ID:    contentBlock.ID,
					Name:  contentBlock.Name,
					Input: input,
				},
			})
		}
	}

	// Ensure message has at least one content element
	// This handles edge cases where LLM returns empty response
	if len(msg.Content) == 0 {
		msg.Content = append(msg.Content, types.Content{
			Type: "text",
			Text: " ", // Single space to satisfy validation
		})
	}

	return msg
}

// convertTextDelta converts Anthropic text delta to types.Content
func convertTextDelta(delta anthropicsdk.TextDelta) types.Content {
	return types.Content{
		Type: "text",
		Text: delta.Text,
	}
}
