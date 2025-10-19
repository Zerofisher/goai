package openai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
	openaisdk "github.com/openai/openai-go/v2"
)

// convertToOpenAIParams converts llm.MessageRequest to OpenAI SDK parameters
func convertToOpenAIParams(req llm.MessageRequest, model string) openaisdk.ChatCompletionNewParams {
	params := openaisdk.ChatCompletionNewParams{
		Model:    openaisdk.ChatModel(model),
		Messages: convertMessages(req.Messages),
	}

	if req.MaxTokens > 0 {
		params.MaxCompletionTokens = openaisdk.Int(int64(req.MaxTokens))
	}

	if req.Temperature > 0 {
		params.Temperature = openaisdk.Float(float64(req.Temperature))
	}

	if req.TopP > 0 {
		params.TopP = openaisdk.Float(float64(req.TopP))
	}

	if len(req.Tools) > 0 {
		params.Tools = convertTools(req.Tools)
	}

	if req.ToolChoice != nil {
		params.ToolChoice = convertToolChoice(req.ToolChoice)
	}

	if req.Seed != nil {
		params.Seed = openaisdk.Int(int64(*req.Seed))
	}

	if len(req.StopSequences) > 0 {
		// OpenAI supports up to 4 stop sequences
		params.Stop = openaisdk.ChatCompletionNewParamsStopUnion{
			OfStringArray: req.StopSequences,
		}
	}

	return params
}

// convertMessages converts []types.Message to OpenAI SDK message format
func convertMessages(messages []types.Message) []openaisdk.ChatCompletionMessageParamUnion {
	result := make([]openaisdk.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			text := msg.GetText()
			result = append(result, openaisdk.UserMessage(text))
		case "assistant":
			toolUses := msg.GetToolUses()
			if len(toolUses) > 0 {
				// Assistant message with tool calls - need to use ToParam from actual message
				// For now, just send text content
				text := msg.GetText()
				result = append(result, openaisdk.AssistantMessage(text))
			} else {
				text := msg.GetText()
				result = append(result, openaisdk.AssistantMessage(text))
			}
		case "system":
			text := msg.GetText()
			result = append(result, openaisdk.SystemMessage(text))
		case "tool":
			// Handle tool result messages
			for _, content := range msg.Content {
				if content.Type == "tool_result" && content.ToolResult != nil {
					result = append(result, openaisdk.ToolMessage(
						content.ToolResult.Content,
						content.ToolResult.ToolUseID,
					))
				}
			}
		}
	}

	return result
}

// convertTools converts tool definitions to OpenAI format
func convertTools(tools []llm.ToolDefinition) []openaisdk.ChatCompletionToolUnionParam {
	result := make([]openaisdk.ChatCompletionToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		result = append(result, openaisdk.ChatCompletionToolUnionParam{
			OfFunction: &openaisdk.ChatCompletionFunctionToolParam{
				Function: openaisdk.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openaisdk.String(tool.Description),
					Parameters:  openaisdk.FunctionParameters(tool.InputSchema),
				},
			},
		})
	}

	return result
}

// convertToolChoice converts tool choice to OpenAI format
func convertToolChoice(choice *llm.ToolChoice) openaisdk.ChatCompletionToolChoiceOptionUnionParam {
	if choice == nil {
		// Auto mode - use default empty struct which means "auto"
		return openaisdk.ChatCompletionToolChoiceOptionUnionParam{}
	}

	switch choice.Type {
	case "auto":
		return openaisdk.ChatCompletionToolChoiceOptionUnionParam{}
	case "none":
		return openaisdk.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openaisdk.String("none"),
		}
	case "any", "required":
		return openaisdk.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openaisdk.String("required"),
		}
	case "tool":
		if choice.ToolName != "" {
			return openaisdk.ToolChoiceOptionFunctionToolChoice(
				openaisdk.ChatCompletionNamedToolChoiceFunctionParam{
					Name: choice.ToolName,
				},
			)
		}
	}

	// Default to auto
	return openaisdk.ChatCompletionToolChoiceOptionUnionParam{}
}

// convertFromOpenAIResponse converts OpenAI response to llm.MessageResponse
func convertFromOpenAIResponse(completion *openaisdk.ChatCompletion) *llm.MessageResponse {
	if len(completion.Choices) == 0 {
		return &llm.MessageResponse{
			ID:    completion.ID,
			Model: string(completion.Model),
		}
	}

	choice := completion.Choices[0]

	resp := &llm.MessageResponse{
		ID:        completion.ID,
		Model:     string(completion.Model),
		Message:   convertAssistantResponseMessage(choice.Message),
		CreatedAt: time.Unix(completion.Created, 0),
	}

	// Check if usage is available
	if completion.Usage.PromptTokens > 0 || completion.Usage.CompletionTokens > 0 {
		resp.Usage = &llm.TokenUsage{
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		}
	}

	return resp
}

// convertAssistantResponseMessage converts OpenAI assistant message to types.Message
func convertAssistantResponseMessage(msg openaisdk.ChatCompletionMessage) types.Message {
	message := types.Message{
		Role:    "assistant",
		Content: []types.Content{},
	}

	// Add text content if present
	if msg.Content != "" {
		message.Content = append(message.Content, types.Content{
			Type: "text",
			Text: msg.Content,
		})
	}

	// Convert tool calls if present
	if len(msg.ToolCalls) > 0 {
		for _, toolCall := range msg.ToolCalls {
			// OpenAI tool calls are always function calls in practice
			// Parse arguments from JSON
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
				// If parsing fails, create an error input
				input = map[string]interface{}{
					"error": fmt.Sprintf("failed to parse arguments: %v", err),
				}
			}

			message.Content = append(message.Content, types.Content{
				Type: "tool_use",
				ToolUse: &types.ToolUse{
					ID:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: input,
				},
			})
		}
	}

	return message
}

// convertStreamChunk converts OpenAI stream chunk to llm.StreamChunk
func convertStreamChunk(chunk openaisdk.ChatCompletionChunk) llm.StreamChunk {
	if len(chunk.Choices) == 0 {
		return llm.StreamChunk{
			ID:    chunk.ID,
			Model: string(chunk.Model),
			Done:  false,
		}
	}

	delta := chunk.Choices[0].Delta
	finishReason := chunk.Choices[0].FinishReason

	streamChunk := llm.StreamChunk{
		ID:    chunk.ID,
		Model: string(chunk.Model),
		Done:  finishReason != "",
	}

	// Add text content if present
	if delta.Content != "" {
		streamChunk.Delta = types.Content{
			Type: "text",
			Text: delta.Content,
		}
	}

	// Add tool calls if present
	// Note: For streaming, we typically get partial tool calls
	// We'll just send the first one if it exists
	if len(delta.ToolCalls) > 0 {
		toolCall := delta.ToolCalls[0]
		var input map[string]interface{}
		if toolCall.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
				input = map[string]interface{}{
					"error": fmt.Sprintf("failed to parse arguments: %v", err),
				}
			}
		}

		streamChunk.Delta = types.Content{
			Type: "tool_use",
			ToolUse: &types.ToolUse{
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: input,
			},
		}
	}

	return streamChunk
}
