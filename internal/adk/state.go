package adk

import (
	"encoding/json"
	"iter"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/memory"
	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/model"
)

// SessionAdapter adapts internal.Session to adk.Session
type SessionAdapter struct {
	sess          *internal.Session
	store         internal.Store
	rootAgentName string
	tokenBudget   int // 0 = use DefaultTokenBudget; set via SessionServiceAdapter
}

func NewSessionAdapter(s *internal.Session, store internal.Store, rootAgentName string) *SessionAdapter {
	return &SessionAdapter{sess: s, store: store, rootAgentName: rootAgentName}
}

func (s *SessionAdapter) ID() string      { return s.sess.Key }
func (s *SessionAdapter) AppName() string { return "lango" }
func (s *SessionAdapter) UserID() string  { return "user" } // Default header

func (s *SessionAdapter) State() session.State {
	return &StateAdapter{sess: s.sess, store: s.store}
}

func (s *SessionAdapter) Events() session.Events {
	return &EventsAdapter{
		history:       s.sess.History,
		tokenBudget:   s.tokenBudget,
		rootAgentName: s.rootAgentName,
	}
}

// EventsWithTokenBudget returns an EventsAdapter that uses token-budget truncation.
func (s *SessionAdapter) EventsWithTokenBudget(budget int) session.Events {
	return &EventsAdapter{
		history:       s.sess.History,
		tokenBudget:   budget,
		rootAgentName: s.rootAgentName,
	}
}

func (s *SessionAdapter) LastUpdateTime() time.Time { return s.sess.UpdatedAt }

// StateAdapter adapts internal.Session.Metadata to adk.State
type StateAdapter struct {
	sess  *internal.Session
	store internal.Store
}

func (s *StateAdapter) Get(key string) (any, error) {
	valStr, ok := s.sess.Metadata[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	var val any
	if err := json.Unmarshal([]byte(valStr), &val); err == nil {
		return val, nil
	}
	return valStr, nil
}

func (s *StateAdapter) Set(key string, val any) error {
	var valStr string
	if sStr, ok := val.(string); ok {
		valStr = sStr
	} else {
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		valStr = string(b)
	}

	if s.sess.Metadata == nil {
		s.sess.Metadata = make(map[string]string)
	}
	s.sess.Metadata[key] = valStr

	return s.store.Update(s.sess)
}

func (s *StateAdapter) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, vStr := range s.sess.Metadata {
			var val any
			if err := json.Unmarshal([]byte(vStr), &val); err != nil {
				val = vStr
			}
			if !yield(k, val) {
				return
			}
		}
	}
}

// DefaultTokenBudget is the token budget used when no explicit budget is provided.
const DefaultTokenBudget = 32000

// ModelTokenBudget returns an appropriate history token budget for the given model.
// It uses approximately 50-60% of each model family's context window, leaving room
// for system prompts, tool definitions, and generated output.
// Returns DefaultTokenBudget for unknown models.
func ModelTokenBudget(modelName string) int {
	lower := strings.ToLower(modelName)
	switch {
	case strings.Contains(lower, "claude"):
		return 100000 // ~50% of 200K context
	case strings.Contains(lower, "gemini"):
		return 200000 // ~20% of 1M context
	case strings.Contains(lower, "gpt-4o"), strings.Contains(lower, "gpt-4-turbo"):
		return 64000 // ~50% of 128K context
	case strings.Contains(lower, "gpt-4"):
		return 32000 // 8K–32K context models
	case strings.Contains(lower, "gpt-3.5"):
		return 8000 // ~50% of 16K context
	default:
		return DefaultTokenBudget
	}
}

// EventsAdapter adapts internal history to adk events.
// Uses token-budget truncation: includes messages from most recent until the budget is exhausted.
// Truncated history and converted events are lazily cached for O(1) repeated access.
type EventsAdapter struct {
	history       []internal.Message
	tokenBudget   int
	rootAgentName string

	// Lazy caches — safe because EventsAdapter is created fresh per session access.
	truncateOnce sync.Once
	truncated    []internal.Message
	eventsOnce   sync.Once
	eventsCache  []*session.Event
}

// truncatedHistory returns the messages to include based on token budget.
// The result is lazily cached so repeated calls (e.g. from Len, At, All) are O(1).
func (e *EventsAdapter) truncatedHistory() []internal.Message {
	e.truncateOnce.Do(func() {
		e.truncated = e.tokenBudgetTruncate()
	})
	return e.truncated
}

// tokenBudgetTruncate includes messages from most recent to oldest until the token budget is exhausted.
// It ensures the truncated slice does not start with a tool/function message or an assistant+FunctionCall
// without its matching tool response, which would violate Gemini's message ordering requirements.
func (e *EventsAdapter) tokenBudgetTruncate() []internal.Message {
	if len(e.history) == 0 {
		return e.history
	}

	budget := e.tokenBudget
	if budget <= 0 {
		budget = DefaultTokenBudget
	}

	var totalTokens int
	startIdx := len(e.history)

	for i := len(e.history) - 1; i >= 0; i-- {
		msgTokens := memory.CountMessageTokens(e.history[i])
		if totalTokens+msgTokens > budget && startIdx < len(e.history) {
			break
		}
		totalTokens += msgTokens
		startIdx = i
	}

	result := e.history[startIdx:]

	// Only apply sequence safety when truncation actually removed messages,
	// because a trailing FunctionCall without response is valid (response pending).
	if startIdx > 0 {
		for len(result) > 0 {
			first := result[0]
			// tool/function messages cannot appear without a preceding assistant+FunctionCall.
			if first.Role == "tool" || first.Role == "function" {
				result = result[1:]
				continue
			}
			// assistant with FunctionCall but no following tool response is invalid at boundary.
			if (first.Role == "assistant" || first.Role == "model") && len(first.ToolCalls) > 0 && hasFunctionCallToolCalls(first.ToolCalls) {
				if len(result) < 2 || (result[1].Role != "tool" && result[1].Role != "function") {
					result = result[1:]
					continue
				}
			}
			break
		}
	}

	return result
}

// hasFunctionCallToolCalls returns true if any ToolCall has Input (i.e., is a FunctionCall, not a FunctionResponse).
func hasFunctionCallToolCalls(tcs []internal.ToolCall) bool {
	for _, tc := range tcs {
		if tc.Input != "" {
			return true
		}
	}
	return false
}

func (e *EventsAdapter) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		msgs := e.truncatedHistory()

		// Track the most recent assistant ToolCalls for legacy tool-message fallback.
		var lastAssistantToolCalls []internal.ToolCall

		// Defense-in-depth: merge consecutive same-role events before yielding.
		// This prevents Gemini INVALID_ARGUMENT errors caused by duplicate
		// model or user turns in the session history.
		var pending *session.Event

		flushPending := func() bool {
			if pending != nil {
				if !yield(pending) {
					return false
				}
				pending = nil
			}
			return true
		}

		for _, msg := range msgs {
			// Map Author: use stored author if available, otherwise derive from role.
			author := msg.Author
			if author == "" {
				switch msg.Role {
				case types.RoleUser:
					author = "user"
				case types.RoleAssistant, types.RoleModel:
					if e.rootAgentName != "" {
						author = e.rootAgentName
					} else {
						author = "lango-agent"
					}
				case types.RoleFunction, types.RoleTool:
					author = "tool"
				default:
					if e.rootAgentName != "" {
						author = e.rootAgentName
					} else {
						author = "lango-agent"
					}
				}
			}

			role := msg.Role
			var parts []*genai.Part

			switch role {
			case types.RoleAssistant, types.RoleModel:
				// Text content
				if msg.Content != "" {
					parts = append(parts, &genai.Part{Text: msg.Content})
				}
				// FunctionCall parts from ToolCalls (only those with Input, not Output-only)
				for _, tc := range msg.ToolCalls {
					if tc.Input == "" {
						continue
					}
					args := make(map[string]any)
					_ = json.Unmarshal([]byte(tc.Input), &args)
					parts = append(parts, &genai.Part{
						FunctionCall: &genai.FunctionCall{
							ID:   tc.ID,
							Name: tc.Name,
							Args: args,
						},
					})
				}
				// Remember for legacy fallback
				lastAssistantToolCalls = msg.ToolCalls

			case types.RoleTool, types.RoleFunction:
				// Check if ToolCalls carry FunctionResponse metadata (new format)
				hasResponseMeta := false
				for _, tc := range msg.ToolCalls {
					if tc.Output != "" {
						hasResponseMeta = true
						break
					}
				}

				if hasResponseMeta {
					// New format: reconstruct FunctionResponse from stored ToolCalls
					for _, tc := range msg.ToolCalls {
						if tc.Output == "" {
							continue
						}
						resp := make(map[string]any)
						_ = json.Unmarshal([]byte(tc.Output), &resp)
						parts = append(parts, &genai.Part{
							FunctionResponse: &genai.FunctionResponse{
								ID:       tc.ID,
								Name:     tc.Name,
								Response: resp,
							},
						})
					}
					role = types.RoleFunction // Gemini expects "function" role for FunctionResponse
				} else if len(lastAssistantToolCalls) > 0 {
					// Legacy format: infer FunctionResponse from preceding assistant ToolCalls.
					// Use the content as the response body for the first call.
					tc := lastAssistantToolCalls[0]
					resp := make(map[string]any)
					if msg.Content != "" {
						_ = json.Unmarshal([]byte(msg.Content), &resp)
						if len(resp) == 0 {
							resp = map[string]any{"result": msg.Content}
						}
					}
					parts = append(parts, &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       tc.ID,
							Name:     tc.Name,
							Response: resp,
						},
					})
					// Consume used call
					if len(lastAssistantToolCalls) > 1 {
						lastAssistantToolCalls = lastAssistantToolCalls[1:]
					} else {
						lastAssistantToolCalls = nil
					}
					role = types.RoleFunction
				} else {
					// No context to reconstruct FunctionResponse — emit as text
					if msg.Content != "" {
						parts = append(parts, &genai.Part{Text: msg.Content})
					}
				}

			default:
				// user or other roles
				if msg.Content != "" {
					parts = append(parts, &genai.Part{Text: msg.Content})
				}
			}

			if len(parts) == 0 {
				continue
			}

			evt := &session.Event{
				ID:        uuid.NewString(),
				Timestamp: msg.Timestamp,
				Author:    author,
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  string(role),
						Parts: parts,
					},
				},
			}

			// Merge consecutive same-role events into a single event.
			if pending != nil && pending.Content.Role == evt.Content.Role {
				pending.Content.Parts = append(pending.Content.Parts, evt.Content.Parts...)
				continue
			}
			if !flushPending() {
				return
			}
			pending = evt
		}
		flushPending()
	}
}

func (e *EventsAdapter) Len() int {
	// Use the cached event list to be consistent with All() and At(),
	// which may merge consecutive same-role events.
	e.eventsOnce.Do(func() {
		for evt := range e.All() {
			e.eventsCache = append(e.eventsCache, evt)
		}
	})
	return len(e.eventsCache)
}

// At returns the i-th event. The full event list is built once on first call
// and cached, making subsequent At calls O(1) instead of O(n).
func (e *EventsAdapter) At(i int) *session.Event {
	e.eventsOnce.Do(func() {
		for evt := range e.All() {
			e.eventsCache = append(e.eventsCache, evt)
		}
	})
	if i < 0 || i >= len(e.eventsCache) {
		return nil
	}
	return e.eventsCache[i]
}
