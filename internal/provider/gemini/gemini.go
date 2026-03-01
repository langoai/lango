package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/langoai/lango/internal/provider"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	id     string
	model  string
}

func NewProvider(ctx context.Context, id, apiKey string, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiProvider{
		client: client,
		id:     id,
		model:  model,
	}, nil
}

func (p *GeminiProvider) ID() string {
	return p.id
}

func (p *GeminiProvider) Generate(ctx context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	// Convert messages to genai.Content
	var contents []*genai.Content
	var systemParts []*genai.Part

	for _, m := range params.Messages {
		if m.Role == "system" {
			systemParts = append(systemParts, &genai.Part{Text: m.Content})
			continue
		}

		if m.Role == "tool" {
			toolCallID, _ := m.Metadata["tool_call_id"].(string)
			toolCallName, _ := m.Metadata["tool_call_name"].(string)

			if toolCallID == "" || toolCallName == "" {
				// Fallback or skip if missing info
				continue
			}

			// Validate response is valid JSON if possible, otherwise wrap it
			// Gemini expects the response to be a structured object in some cases,
			// or a simple map.
			// m.Content is the result string (JSON).
			var responseContent map[string]interface{}
			if err := json.Unmarshal([]byte(m.Content), &responseContent); err != nil {
				// If not JSON object, wrap it
				responseContent = map[string]interface{}{"result": m.Content}
			}

			contents = append(contents, &genai.Content{
				Role: "user", // Must be user or model
				Parts: []*genai.Part{
					{
						FunctionResponse: &genai.FunctionResponse{
							Name:     toolCallName,
							Response: responseContent,
						},
					},
				},
			})
			continue
		}

		role := m.Role
		if role == "assistant" {
			role = "model"
		}

		var parts []*genai.Part
		if m.Content != "" {
			parts = append(parts, &genai.Part{Text: m.Content})
		}

		// If assistant message has tool calls, add them as parts
		if role == "model" && len(m.ToolCalls) > 0 {
			// If content is empty/present, we append ToolCalls
			// Note: Gemini can have text AND parts.
			for _, tc := range m.ToolCalls {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
					args = make(map[string]interface{})
				}
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: tc.Name,
						Args: args,
					},
				})
			}
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}

	// Tools
	var tools []*genai.Tool
	if len(params.Tools) > 0 {
		var funcDecls []*genai.FunctionDeclaration
		for _, t := range params.Tools {
			schema, err := convertSchema(t.Parameters)
			if err != nil {
				return nil, fmt.Errorf("convert tool schema: %w", err)
			}
			funcDecls = append(funcDecls, &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			})
		}
		tools = append(tools, &genai.Tool{FunctionDeclarations: funcDecls})
	}

	model := p.model
	if params.Model != "" {
		model = params.Model
	}

	// Alias "gemini" to a valid model
	if model == "gemini" {
		model = "gemini-3-flash-preview"
	}

	temp := float32(params.Temperature)
	maxTokens := int32(params.MaxTokens)

	conf := &genai.GenerateContentConfig{
		Temperature:     &temp,
		MaxOutputTokens: maxTokens,
		Tools:           tools,
	}

	if len(systemParts) > 0 {
		conf.SystemInstruction = &genai.Content{
			Parts: systemParts,
		}
	}

	// Sanitize contents to satisfy Gemini's strict turn-ordering rules
	// (no consecutive same-role turns, FunctionCall/Response pairing, etc).
	contents = sanitizeContents(contents)

	// Streaming
	streamIter := p.client.Models.GenerateContentStream(ctx, model, contents, conf)

	return func(yield func(provider.StreamEvent, error) bool) {
		for resp, err := range streamIter {
			if err != nil {
				yield(provider.StreamEvent{Type: provider.StreamEventError, Error: err}, err)
				return
			}

			// Handle response parts
			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						if part.Text != "" {
							if !yield(provider.StreamEvent{
								Type: provider.StreamEventPlainText,
								Text: part.Text,
							}, nil) {
								return
							}
						}
						if part.FunctionCall != nil {
							argsJSON, _ := json.Marshal(part.FunctionCall.Args)
							if !yield(provider.StreamEvent{
								Type: provider.StreamEventToolCall,
								ToolCall: &provider.ToolCall{
									ID:        part.FunctionCall.Name, // Use name as ID if ID missing
									Name:      part.FunctionCall.Name,
									Arguments: string(argsJSON),
								},
							}, nil) {
								return
							}
						}
					}
				}
			}
		}
		yield(provider.StreamEvent{Type: provider.StreamEventDone}, nil)
	}, nil
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]provider.ModelInfo, error) {
	var models []provider.ModelInfo
	for m, err := range p.client.Models.All(ctx) {
		if err != nil {
			if len(models) > 0 {
				return models, nil
			}
			return nil, fmt.Errorf("list gemini models: %w", err)
		}
		id := strings.TrimPrefix(m.Name, "models/")
		models = append(models, provider.ModelInfo{
			ID:            id,
			Name:          m.DisplayName,
			ContextWindow: int(m.InputTokenLimit),
		})
	}
	return models, nil
}

func convertSchema(schemaMap map[string]interface{}) (*genai.Schema, error) {
	b, err := json.Marshal(schemaMap)
	if err != nil {
		return nil, err
	}
	var s genai.Schema
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
