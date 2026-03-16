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
	msgs := make([]openai.ChatCompletionMessage, len(params.Messages))
	for i, m := range params.Messages {
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
				},
			})
		}
		if len(tools) > 0 {
			req.Tools = tools
		}
	}

	return req, nil
}
