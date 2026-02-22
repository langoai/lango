package adk

import (
	"context"
	"encoding/json"
	"iter"
	"strings"

	"github.com/langoai/lango/internal/provider"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type ModelAdapter struct {
	p     provider.Provider
	model string
}

func NewModelAdapter(p provider.Provider, model string) *ModelAdapter {
	return &ModelAdapter{p: p, model: model}
}

func (m *ModelAdapter) Name() string {
	return m.model
}

func (m *ModelAdapter) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		msgs, err := convertMessages(req.Contents)
		if err != nil {
			yield(nil, err)
			return
		}

		tools, err := convertTools(req.Config)
		if err != nil {
			yield(nil, err)
			return
		}

		// Forward ADK system instruction as a system message for the provider.
		if req.Config != nil && req.Config.SystemInstruction != nil {
			sysText := extractSystemText(req.Config.SystemInstruction)
			if sysText != "" {
				sysMsg := provider.Message{Role: "system", Content: sysText}
				msgs = append([]provider.Message{sysMsg}, msgs...)
			}
		}

		params := provider.GenerateParams{
			Model:    req.Model,
			Messages: msgs,
			Tools:    tools,
		}

		if req.Config != nil {
			if req.Config.Temperature != nil {
				params.Temperature = float64(*req.Config.Temperature)
			}
			if req.Config.MaxOutputTokens != 0 {
				params.MaxTokens = int(req.Config.MaxOutputTokens)
			}
		}
		if params.Model == "" {
			// Fallback if not set in request (ADK might set it in client/factory)
			// But params must have it.
			// We can default or error.
			// provider usually requires it.
		}

		pSeq, err := m.p.Generate(ctx, params)
		if err != nil {
			yield(nil, err)
			return
		}

		if stream {
			// Streaming mode: yield partial text events for real-time UI,
			// and include accumulated full text in the final done event
			// so the ADK runner stores the complete response in the session.
			var accumulated strings.Builder
			var toolParts []*genai.Part

			for evt, err := range pSeq {
				if err != nil {
					yield(nil, err)
					return
				}

				switch evt.Type {
				case provider.StreamEventPlainText:
					accumulated.WriteString(evt.Text)
					resp := &model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{{Text: evt.Text}},
						},
						Partial: true,
					}
					if !yield(resp, nil) {
						return
					}

				case provider.StreamEventToolCall:
					if evt.ToolCall != nil {
						args := make(map[string]any)
						_ = json.Unmarshal([]byte(evt.ToolCall.Arguments), &args)
						part := &genai.Part{
							FunctionCall: &genai.FunctionCall{
								Name: evt.ToolCall.Name,
								Args: args,
							},
						}
						toolParts = append(toolParts, part)
						resp := &model.LLMResponse{
							Content: &genai.Content{
								Role:  "model",
								Parts: []*genai.Part{part},
							},
						}
						if !yield(resp, nil) {
							return
						}
					}

				case provider.StreamEventDone:
					// Final event: include accumulated full text so ADK
					// stores a complete assistant message in the session.
					var finalParts []*genai.Part
					if text := accumulated.String(); text != "" {
						finalParts = append(finalParts, &genai.Part{Text: text})
					}
					finalParts = append(finalParts, toolParts...)
					resp := &model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: finalParts,
						},
						TurnComplete: true,
						Partial:      false,
					}
					if !yield(resp, nil) {
						return
					}

				case provider.StreamEventError:
					yield(nil, evt.Error)
					return
				}
			}
		} else {
			// Non-streaming mode: accumulate all events internally and
			// yield a single complete response for session storage.
			var textAccum strings.Builder
			var toolParts []*genai.Part

			for evt, err := range pSeq {
				if err != nil {
					yield(nil, err)
					return
				}

				switch evt.Type {
				case provider.StreamEventPlainText:
					textAccum.WriteString(evt.Text)
				case provider.StreamEventToolCall:
					if evt.ToolCall != nil {
						args := make(map[string]any)
						_ = json.Unmarshal([]byte(evt.ToolCall.Arguments), &args)
						toolParts = append(toolParts, &genai.Part{
							FunctionCall: &genai.FunctionCall{
								Name: evt.ToolCall.Name,
								Args: args,
							},
						})
					}
				case provider.StreamEventDone:
					// Ignored — we build the final response below.
				case provider.StreamEventError:
					yield(nil, evt.Error)
					return
				}
			}

			var parts []*genai.Part
			if text := textAccum.String(); text != "" {
				parts = append(parts, &genai.Part{Text: text})
			}
			parts = append(parts, toolParts...)

			yield(&model.LLMResponse{
				Content:      &genai.Content{Role: "model", Parts: parts},
				TurnComplete: true,
				Partial:      false,
			}, nil)
		}
	}
}

func convertMessages(contents []*genai.Content) ([]provider.Message, error) {
	var msgs []provider.Message
	for _, c := range contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		} else if role == "function" {
			role = "tool"
		}

		msg := provider.Message{Role: role}
		for _, p := range c.Parts {
			if p.Text != "" {
				msg.Content += p.Text
			}
			if p.FunctionCall != nil {
				b, _ := json.Marshal(p.FunctionCall.Args)
				id := p.FunctionCall.ID
				if id == "" {
					id = "call_" + p.FunctionCall.Name
				}
				msg.ToolCalls = append(msg.ToolCalls, provider.ToolCall{
					ID:        id,
					Name:      p.FunctionCall.Name,
					Arguments: string(b),
				})
			}
			if p.FunctionResponse != nil {
				b, _ := json.Marshal(p.FunctionResponse.Response)
				msg.Content += string(b)
				if msg.Metadata == nil {
					msg.Metadata = make(map[string]interface{})
				}
				id := p.FunctionResponse.ID
				if id == "" {
					id = p.FunctionResponse.Name
				}
				msg.Metadata["tool_call_id"] = id
			}
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// extractSystemText concatenates all text parts from a genai.Content into a single string.
func extractSystemText(content *genai.Content) string {
	var parts []string
	for _, p := range content.Parts {
		if p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

func convertTools(cfg *genai.GenerateContentConfig) ([]provider.Tool, error) {
	var tools []provider.Tool
	if cfg == nil || cfg.Tools == nil {
		return tools, nil
	}

	for _, t := range cfg.Tools {
		if t.FunctionDeclarations != nil {
			for _, fd := range t.FunctionDeclarations {
				// Convert Schema to map
				schemaMap := make(map[string]interface{})
				if fd.Parameters != nil {
					// genai.Schema to map is complex if we recurse.
					// But we can json marshal/unmarshal
					b, err := json.Marshal(fd.Parameters)
					if err == nil {
						_ = json.Unmarshal(b, &schemaMap)
					}
				}

				tools = append(tools, provider.Tool{
					Name:        fd.Name,
					Description: fd.Description,
					Parameters:  schemaMap,
				})
			}
		}
	}
	return tools, nil
}
