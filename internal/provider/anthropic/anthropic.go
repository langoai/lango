package anthropic

import (
	"context"
	"iter"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/provider"
)

var logger = logging.SubsystemSugar("provider.anthropic")

type AnthropicProvider struct {
	client *anthropic.Client
	id     string
}

func NewProvider(id, apiKey string) *AnthropicProvider {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicProvider{
		client: &client,
		id:     id,
	}
}

func (p *AnthropicProvider) ID() string {
	return p.id
}

func (p *AnthropicProvider) Generate(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	msgParams, err := p.convertParams(params)
	if err != nil {
		return nil, err
	}

	stream := p.client.Messages.NewStreaming(ctx, msgParams)

	return func(yield func(provider.StreamEvent, error) bool) {
		for stream.Next() {
			evt := stream.Current()

			switch evt.Type {
			case "content_block_delta":
				if evt.Delta.Type == "text_delta" {
					if !yield(provider.StreamEvent{
						Type: provider.StreamEventPlainText,
						Text: evt.Delta.Text,
					}, nil) {
						return
					}
				} else if evt.Delta.Type == "input_json_delta" {
					if !yield(provider.StreamEvent{
						Type: provider.StreamEventToolCall,
						ToolCall: &provider.ToolCall{
							Arguments: evt.Delta.PartialJSON,
						},
					}, nil) {
						return
					}
				}
			case "content_block_start":
				if evt.ContentBlock.Type == "tool_use" {
					if !yield(provider.StreamEvent{
						Type: provider.StreamEventToolCall,
						ToolCall: &provider.ToolCall{
							ID:   evt.ContentBlock.ID,
							Name: evt.ContentBlock.Name,
						},
					}, nil) {
						return
					}
				}
			case "message_delta":
				// Handle stop reason if needed
			}
		}

		if err := stream.Err(); err != nil {
			yield(provider.StreamEvent{Type: provider.StreamEventError, Error: err}, err)
			return
		}

		yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
	}, nil
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]provider.ModelInfo, error) {
	return []provider.ModelInfo{
		{ID: "claude-3-5-sonnet-latest", Name: "Claude 3.5 Sonnet"},
		{ID: "claude-3-opus-latest", Name: "Claude 3 Opus"},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku"},
	}, nil
}

func (p *AnthropicProvider) convertParams(params provider.GenerateParams) (anthropic.MessageNewParams, error) {
	var msgs []anthropic.MessageParam
	for _, m := range params.Messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		case "system":
			// system role handled separately below
		default:
			logger.Warnw("unknown message role, skipping", "role", m.Role)
		}
	}

	req := anthropic.MessageNewParams{
		Model:     anthropic.Model(params.Model),
		Messages:  msgs,
		MaxTokens: int64(params.MaxTokens),
	}

	for _, m := range params.Messages {
		if m.Role == "system" {
			req.System = []anthropic.TextBlockParam{{
				Text: m.Content,
				Type: constant.Text("text"),
			}}
			break
		}
	}

	if params.Temperature > 0 {
		req.Temperature = param.NewOpt(params.Temperature)
	}

	if len(params.Tools) > 0 {
		var tools []anthropic.ToolUnionParam
		for _, t := range params.Tools {
			var required []string
			if reqRaw, ok := t.Parameters["required"]; ok {
				if reqSlice, ok := reqRaw.([]interface{}); ok {
					for _, r := range reqSlice {
						if s, ok := r.(string); ok {
							required = append(required, s)
						}
					}
				} else if reqSlice, ok := reqRaw.([]string); ok {
					required = reqSlice
				}
			}

			schema := anthropic.ToolInputSchemaParam{
				Type:       constant.Object("object"),
				Properties: t.Parameters["properties"],
				Required:   required,
			}

			tool := anthropic.ToolUnionParamOfTool(schema, t.Name)
			if t.Description != "" {
				tool.OfTool.Description = param.NewOpt(t.Description)
			}
			tools = append(tools, tool)
		}
		req.Tools = tools
	}

	return req, nil
}
