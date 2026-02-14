package adk

import (
	"encoding/json"
	"iter"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/langowarny/lango/internal/memory"
	internal "github.com/langowarny/lango/internal/session"
	"google.golang.org/adk/model"
)

// SessionAdapter adapts internal.Session to adk.Session
type SessionAdapter struct {
	sess  *internal.Session
	store internal.Store
}

func NewSessionAdapter(s *internal.Session, store internal.Store) *SessionAdapter {
	return &SessionAdapter{sess: s, store: store}
}

func (s *SessionAdapter) ID() string      { return s.sess.Key }
func (s *SessionAdapter) AppName() string { return "lango" }
func (s *SessionAdapter) UserID() string  { return "user" } // Default header

func (s *SessionAdapter) State() session.State {
	return &StateAdapter{sess: s.sess, store: s.store}
}

func (s *SessionAdapter) Events() session.Events {
	return &EventsAdapter{history: s.sess.History}
}

// EventsWithTokenBudget returns an EventsAdapter that uses token-budget truncation.
func (s *SessionAdapter) EventsWithTokenBudget(budget int) session.Events {
	return &EventsAdapter{
		history:     s.sess.History,
		tokenBudget: budget,
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

// EventsAdapter adapts internal history to adk events.
// Supports two truncation modes:
// - Token-budget: when tokenBudget > 0, includes messages from most recent until budget is exhausted
// - Message count: falls back to hard 100-message cap when tokenBudget is 0
type EventsAdapter struct {
	history     []internal.Message
	tokenBudget int // 0 means use legacy 100-message cap
}

// truncatedHistory returns the messages to include based on truncation mode.
func (e *EventsAdapter) truncatedHistory() []internal.Message {
	if e.tokenBudget > 0 {
		return e.tokenBudgetTruncate()
	}
	// Legacy: hard 100-message cap
	msgs := e.history
	if len(msgs) > 100 {
		msgs = msgs[len(msgs)-100:]
	}
	return msgs
}

// tokenBudgetTruncate includes messages from most recent to oldest until the token budget is exhausted.
func (e *EventsAdapter) tokenBudgetTruncate() []internal.Message {
	if len(e.history) == 0 {
		return e.history
	}

	var totalTokens int
	startIdx := len(e.history)

	for i := len(e.history) - 1; i >= 0; i-- {
		msgTokens := memory.CountMessageTokens(e.history[i])
		if totalTokens+msgTokens > e.tokenBudget && startIdx < len(e.history) {
			break
		}
		totalTokens += msgTokens
		startIdx = i
	}

	return e.history[startIdx:]
}

func (e *EventsAdapter) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		msgs := e.truncatedHistory()

		for _, msg := range msgs {
			// Map Author
			author := "model" // Default for assistant
			if msg.Role == "user" {
				author = "user"
			} else if msg.Role == "assistant" {
				author = "lango-agent" // ADK agent name
			} else if msg.Role == "function" || msg.Role == "tool" {
				author = "tool"
			}

			var parts []*genai.Part
			if msg.Content != "" {
				parts = append(parts, &genai.Part{Text: msg.Content})
			}

			evt := &session.Event{
				ID:        uuid.NewString(), // Generate on fly as we don't store event IDs
				Timestamp: msg.Timestamp,
				Author:    author,
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  msg.Role,
						Parts: parts,
					},
				},
			}
			// Map tool calls if present
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					args := make(map[string]any)
					// best effort json unmarshal
					_ = json.Unmarshal([]byte(tc.Input), &args)

					evt.LLMResponse.Content.Parts = append(evt.LLMResponse.Content.Parts, &genai.Part{
						FunctionCall: &genai.FunctionCall{
							Name: tc.Name,
							Args: args,
						},
					})
				}
			}
			if !yield(evt) {
				return
			}
		}
	}
}

func (e *EventsAdapter) Len() int {
	return len(e.truncatedHistory())
}

func (e *EventsAdapter) At(i int) *session.Event {
	// Reusing logic from All is inefficient but simple for now
	var found *session.Event
	count := 0
	e.All()(func(evt *session.Event) bool {
		if count == i {
			found = evt
			return false
		}
		count++
		return true
	})
	return found
}
