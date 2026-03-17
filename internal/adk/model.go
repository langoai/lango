package adk

import (
	"context"
	"encoding/json"
	"iter"
	"sort"
	"strings"

	"github.com/langoai/lango/internal/provider"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// accumEntry holds the accumulated state for a single tool call being assembled
// from streaming chunks.
type accumEntry struct {
	index            int
	id               string
	name             string
	args             strings.Builder
	thought          bool
	thoughtSignature []byte
}

// toolCallAccumulator assembles streaming tool call deltas into complete FunctionCall parts.
// It supports both OpenAI (Index-based correlation) and Anthropic (ID/Name start + orphan delta)
// streaming patterns.
type toolCallAccumulator struct {
	entries   map[int]*accumEntry
	nextIndex int // auto-increment for entries without explicit Index
	lastIndex int // last active entry for orphan deltas
	hasAny    bool
}

func (a *toolCallAccumulator) add(tc *provider.ToolCall) {
	if tc == nil {
		return
	}
	if a.entries == nil {
		a.entries = make(map[int]*accumEntry)
	}

	// Resolve entry index via fallback chain.
	var idx int
	switch {
	case tc.Index != nil:
		// OpenAI: explicit chunk correlation index
		idx = *tc.Index
	case tc.ID != "" || tc.Name != "":
		// Anthropic start: new tool call, assign synthetic index
		idx = a.nextIndex
		a.nextIndex++
	default:
		// Anthropic delta / orphan: append to last active entry
		if !a.hasAny {
			logger().Warnw("dropping orphan tool call delta", "args_len", len(tc.Arguments))
			return
		}
		idx = a.lastIndex
	}

	if _, exists := a.entries[idx]; !exists {
		a.entries[idx] = &accumEntry{index: idx}
	}
	a.lastIndex = idx
	a.hasAny = true

	e := a.entries[idx]
	if tc.ID != "" {
		e.id = tc.ID
	}
	if tc.Name != "" {
		e.name = tc.Name
	}
	if tc.Arguments != "" {
		e.args.WriteString(tc.Arguments)
	}
	if tc.Thought {
		e.thought = true
	}
	if len(tc.ThoughtSignature) > 0 {
		e.thoughtSignature = tc.ThoughtSignature
	}
}

func (a *toolCallAccumulator) done() []*genai.Part {
	// Sort by index for deterministic output.
	indices := make([]int, 0, len(a.entries))
	for idx := range a.entries {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	parts := make([]*genai.Part, 0, len(indices))
	for _, idx := range indices {
		e := a.entries[idx]
		if e.name == "" {
			logger().Warnw("dropping accumulated tool call with empty name", "index", idx, "id", e.id)
			continue
		}
		args := make(map[string]any)
		if raw := e.args.String(); raw != "" {
			_ = json.Unmarshal([]byte(raw), &args)
		}
		id := e.id
		if id == "" {
			id = "call_" + e.name
		}
		parts = append(parts, &genai.Part{
			FunctionCall: &genai.FunctionCall{
				ID:   id,
				Name: e.name,
				Args: args,
			},
			Thought:          e.thought,
			ThoughtSignature: e.thoughtSignature,
		})
	}
	return parts
}

// TokenUsageCallback is called when a provider returns token usage data.
type TokenUsageCallback func(providerID, model string, input, output, total, cache int64)

type ModelAdapter struct {
	p            provider.Provider
	model        string
	OnTokenUsage TokenUsageCallback
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
		// params.Model may be empty here; the provider will use its default.

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
			var toolAccum toolCallAccumulator

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
					toolAccum.add(evt.ToolCall)
					// Yield partial tool call notification only when Name is present
					// (first chunk with identity). Subsequent arg-only deltas are
					// accumulated silently to avoid storing incomplete FunctionCalls.
					if evt.ToolCall != nil && evt.ToolCall.Name != "" {
						args := make(map[string]any)
						if evt.ToolCall.Arguments != "" {
							_ = json.Unmarshal([]byte(evt.ToolCall.Arguments), &args)
						}
						id := evt.ToolCall.ID
						if id == "" {
							id = "call_" + evt.ToolCall.Name
						}
						part := &genai.Part{
							FunctionCall: &genai.FunctionCall{
								ID:   id,
								Name: evt.ToolCall.Name,
								Args: args,
							},
							Thought:          evt.ToolCall.Thought,
							ThoughtSignature: evt.ToolCall.ThoughtSignature,
						}
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

				case provider.StreamEventThought:
					// Thought text filtered at provider level; no action needed.

				case provider.StreamEventDone:
					// Forward token usage to callback if available.
					if evt.Usage != nil && m.OnTokenUsage != nil {
						m.OnTokenUsage(m.p.ID(), m.model, evt.Usage.InputTokens, evt.Usage.OutputTokens, evt.Usage.TotalTokens, evt.Usage.CacheTokens)
					}

					// Final event: include accumulated full text and fully
					// assembled tool calls so ADK stores a complete assistant
					// message in the session.
					var finalParts []*genai.Part
					if text := accumulated.String(); text != "" {
						finalParts = append(finalParts, &genai.Part{Text: text})
					}
					finalParts = append(finalParts, toolAccum.done()...)
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
			var toolAccum toolCallAccumulator

			for evt, err := range pSeq {
				if err != nil {
					yield(nil, err)
					return
				}

				switch evt.Type {
				case provider.StreamEventPlainText:
					textAccum.WriteString(evt.Text)
				case provider.StreamEventToolCall:
					toolAccum.add(evt.ToolCall)
				case provider.StreamEventThought:
					// Thought text filtered at provider level; no action needed.
				case provider.StreamEventDone:
					// Forward token usage to callback if available.
					if evt.Usage != nil && m.OnTokenUsage != nil {
						m.OnTokenUsage(m.p.ID(), m.model, evt.Usage.InputTokens, evt.Usage.OutputTokens, evt.Usage.TotalTokens, evt.Usage.CacheTokens)
					}
				case provider.StreamEventError:
					yield(nil, evt.Error)
					return
				}
			}

			var parts []*genai.Part
			if text := textAccum.String(); text != "" {
				parts = append(parts, &genai.Part{Text: text})
			}
			parts = append(parts, toolAccum.done()...)

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
		switch role {
		case "model":
			role = "assistant"
		case "function":
			role = "tool"
		}

		msg := provider.Message{Role: role}
		for _, p := range c.Parts {
			if p.Text != "" {
				msg.Content += p.Text
			}
			if p.FunctionCall != nil {
				if p.FunctionCall.Name == "" {
					logger().Warnw("skipping FunctionCall with empty name", "role", c.Role, "id", p.FunctionCall.ID)
					continue
				}
				b, _ := json.Marshal(p.FunctionCall.Args)
				id := p.FunctionCall.ID
				if id == "" {
					id = "call_" + p.FunctionCall.Name
				}
				msg.ToolCalls = append(msg.ToolCalls, provider.ToolCall{
					ID:               id,
					Name:             p.FunctionCall.Name,
					Arguments:        string(b),
					Thought:          p.Thought,
					ThoughtSignature: p.ThoughtSignature,
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
				msg.Metadata["tool_call_name"] = p.FunctionResponse.Name
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
				if fd.Name == "" {
					logger().Warnw("skipping FunctionDeclaration with empty name")
					continue
				}
				// Convert schema to map. ADK v0.5.0+ uses ParametersJsonSchema
				// (a *jsonschema.Schema), while legacy tools use Parameters (*genai.Schema).
				schemaMap := make(map[string]interface{})
				switch {
				case fd.ParametersJsonSchema != nil:
					b, err := json.Marshal(fd.ParametersJsonSchema)
					if err == nil {
						_ = json.Unmarshal(b, &schemaMap)
					}
				case fd.Parameters != nil:
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
