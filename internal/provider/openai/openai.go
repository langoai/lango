package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/provider"
)

var logger = logging.SubsystemSugar("provider.openai")

// OpenAIProvider implements the Provider interface for OpenAI-compatible APIs.
type OpenAIProvider struct {
	client *openai.Client
	id     string
}

// NewProvider creates a new OpenAIProvider.
func NewProvider(id, apiKey, baseURL string) *OpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return &OpenAIProvider{
		client: openai.NewClientWithConfig(config),
		id:     id,
	}
}

// ID returns the provider ID.
func (p *OpenAIProvider) ID() string {
	return p.id
}

// Generate streams responses for the given conversation.
func (p *OpenAIProvider) Generate(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	req, err := p.convertParams(params)
	if err != nil {
		return nil, err
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "does not support tools") {
			return nil, fmt.Errorf("provider error: model '%s' does not support tools. Please try a different model (e.g., llama3, mistral-nemo, or qwen2.5)", params.Model)
		}
		return nil, err
	}

	return func(yield func(provider.StreamEvent, error) bool) {
		defer stream.Close()

		var usage *provider.Usage
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				yield(provider.StreamEvent{Type: provider.StreamEventDone, Usage: usage}, nil)
				return
			}
			if err != nil {
				yield(provider.StreamEvent{Type: provider.StreamEventError, Error: err}, err)
				return
			}

			if len(response.Choices) == 0 {
				// Usage chunk: last chunk with IncludeUsage has no choices but has Usage.
				if response.Usage != nil {
					usage = &provider.Usage{
						InputTokens:  int64(response.Usage.PromptTokens),
						OutputTokens: int64(response.Usage.CompletionTokens),
						TotalTokens:  int64(response.Usage.TotalTokens),
					}
				}
				continue
			}
			delta := response.Choices[0].Delta

			// Handle text content
			if delta.Content != "" {
				if !yield(provider.StreamEvent{
					Type: provider.StreamEventPlainText,
					Text: delta.Content,
				}, nil) {
					return
				}
			}

			// Handle tool calls
			if len(delta.ToolCalls) > 0 {
				for _, tc := range delta.ToolCalls {
					if !yield(provider.StreamEvent{
						Type: provider.StreamEventToolCall,
						ToolCall: &provider.ToolCall{
							Index:     tc.Index,
							ID:        tc.ID,
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}, nil) {
						return
					}
				}
			}
		}
	}, nil
}

// ListModels returns a list of available models.
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]provider.ModelInfo, error) {
	logger.Debugw("listing models", "provider", p.id)
	list, err := p.client.ListModels(ctx)
	if err != nil {
		logger.Debugw("list models failed", "provider", p.id, "error", err)
		return nil, err
	}

	models := make([]provider.ModelInfo, 0, len(list.Models))
	for _, m := range list.Models {
		models = append(models, provider.ModelInfo{
			ID:   m.ID,
			Name: m.ID,
		})
	}
	logger.Debugw("list models succeeded", "provider", p.id, "count", len(models))
	return models, nil
}

func (p *OpenAIProvider) convertParams(params provider.GenerateParams) (openai.ChatCompletionRequest, error) {
	repairedMsgs := repairOrphanedToolCalls(params.Messages)
	msgs := make([]openai.ChatCompletionMessage, len(repairedMsgs))
	for i, m := range repairedMsgs {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if len(m.ToolCalls) > 0 {
			tcs := make([]openai.ToolCall, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				if tc.Name == "" {
					logger.Warnw("filtering tool call with empty name", "id", tc.ID)
					continue
				}
				tcs = append(tcs, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
			if len(tcs) > 0 {
				msg.ToolCalls = tcs
			}
		}
		if toolCallID, ok := m.Metadata["tool_call_id"].(string); ok {
			msg.ToolCallID = toolCallID
		}
		msgs[i] = msg
	}

	req := openai.ChatCompletionRequest{
		Model:         params.Model,
		Messages:      msgs,
		MaxTokens:     params.MaxTokens,
		Temperature:   float32(params.Temperature),
		Stream:        true,
		StreamOptions: &openai.StreamOptions{IncludeUsage: true},
	}

	if len(params.Tools) > 0 {
		tools := make([]openai.Tool, 0, len(params.Tools))
		for _, t := range params.Tools {
			if t.Name == "" {
				logger.Warnw("filtering tool with empty name")
				continue
			}
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
					Strict:      canUseStrictMode(t.Parameters),
				},
			})
		}
		if len(tools) > 0 {
			req.Tools = tools
		}
	}

	return req, nil
}

// canUseStrictMode returns true when a tool's parameter schema satisfies OpenAI's
// strict mode requirements: additionalProperties must be false, and every
// declared property must be listed in "required".
func canUseStrictMode(params map[string]interface{}) bool {
	if params == nil {
		return false
	}
	// Must have additionalProperties: false.
	ap, ok := params["additionalProperties"]
	if !ok {
		return false
	}
	if apBool, ok := ap.(bool); !ok || apBool {
		return false
	}
	// All properties must be in required.
	propsRaw, ok := params["properties"]
	if !ok {
		return true // no properties, nothing to require
	}
	propsMap, ok := propsRaw.(map[string]interface{})
	if !ok {
		return false
	}
	if len(propsMap) == 0 {
		return true
	}
	reqRaw, ok := params["required"]
	if !ok {
		return false // has properties but no required
	}
	reqSet := make(map[string]bool)
	switch req := reqRaw.(type) {
	case []string:
		for _, r := range req {
			reqSet[r] = true
		}
	case []interface{}:
		for _, r := range req {
			if s, ok := r.(string); ok {
				reqSet[s] = true
			}
		}
	default:
		return false
	}
	for name := range propsMap {
		if !reqSet[name] {
			return false
		}
	}
	return true
}

// repairOrphanedToolCalls injects synthetic error responses when an assistant
// tool call is followed by a non-tool message without a matching tool response.
// OpenAI API returns 400 without this repair. Pending calls at history end are untouched.
func repairOrphanedToolCalls(msgs []provider.Message) []provider.Message {
	var result []provider.Message
	for i, msg := range msgs {
		result = append(result, msg)
		if msg.Role != "assistant" || len(msg.ToolCalls) == 0 {
			continue
		}
		// Scan forward: check whether each tool call has a matching tool response
		// before the next non-tool message.
		answered := make(map[string]bool, len(msg.ToolCalls))
		hasFollowingUser := false
		for j := i + 1; j < len(msgs); j++ {
			if msgs[j].Role == "tool" {
				if id, ok := msgs[j].Metadata["tool_call_id"].(string); ok {
					answered[id] = true
				}
				continue
			}
			// Non-tool message (user, assistant, etc.) — orphan boundary
			hasFollowingUser = true
			break
		}
		// Pending calls at end of history are valid (response pending)
		if !hasFollowingUser {
			continue
		}
		// Inject synthetic response only for unanswered calls
		for _, tc := range msg.ToolCalls {
			if tc.ID != "" && !answered[tc.ID] {
				logger.Warnw("injecting synthetic tool response for orphaned tool call",
					"call_id", tc.ID, "name", tc.Name)
				result = append(result, provider.Message{
					Role:    "tool",
					Content: `{"error":"tool call was interrupted and did not complete"}`,
					Metadata: map[string]interface{}{
						"tool_call_id": tc.ID,
					},
				})
			}
		}
	}
	return result
}
