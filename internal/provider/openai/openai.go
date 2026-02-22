package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/langoai/lango/internal/provider"
)

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

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
				return
			}
			if err != nil {
				yield(provider.StreamEvent{Type: provider.StreamEventError, Error: err}, err)
				return
			}

			if len(response.Choices) == 0 {
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
							ID: tc.ID,
							// Usually ID is string.
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
	list, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var models []provider.ModelInfo
	for _, m := range list.Models {
		models = append(models, provider.ModelInfo{
			ID:   m.ID,
			Name: m.ID,
		})
	}
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
			tcs := make([]openai.ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				tcs[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
			msg.ToolCalls = tcs
		}
		if toolCallID, ok := m.Metadata["tool_call_id"].(string); ok {
			msg.ToolCallID = toolCallID
		}
		msgs[i] = msg
	}

	req := openai.ChatCompletionRequest{
		Model:       params.Model,
		Messages:    msgs,
		MaxTokens:   params.MaxTokens,
		Temperature: float32(params.Temperature),
		Stream:      true,
	}

	if len(params.Tools) > 0 {
		tools := make([]openai.Tool, len(params.Tools))
		for i, t := range params.Tools {
			// Parameters should be map[string]interface{}
			// We need to marshal it to match openai-go structure expectation or just assign if compatible
			// sashabaranov/go-openai expects FunctionDefinition.Parameters to be interface{} (usually map or json.RawMessage)

			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			}
		}
		req.Tools = tools
	}

	return req, nil
}
