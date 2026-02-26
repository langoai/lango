package adk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/session"
)

type SessionServiceAdapter struct {
	store         internal.Store
	rootAgentName string
	tokenBudget   int // 0 = use DefaultTokenBudget
}

func NewSessionServiceAdapter(store internal.Store, rootAgentName string) *SessionServiceAdapter {
	return &SessionServiceAdapter{store: store, rootAgentName: rootAgentName}
}

// WithTokenBudget sets the token budget for history truncation.
// Use ModelTokenBudget(modelName) to derive an appropriate budget from the model name.
func (s *SessionServiceAdapter) WithTokenBudget(budget int) *SessionServiceAdapter {
	s.tokenBudget = budget
	return s
}

func (s *SessionServiceAdapter) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	// Create new internal session
	sess := &internal.Session{
		Key:       req.SessionID,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.State != nil {
		for k, v := range req.State {
			var valStr string
			if sStr, ok := v.(string); ok {
				valStr = sStr
			} else {
				b, _ := json.Marshal(v)
				valStr = string(b)
			}
			sess.Metadata[k] = valStr
		}
	}

	if err := s.store.Create(sess); err != nil {
		return nil, err
	}

	sa := NewSessionAdapter(sess, s.store, s.rootAgentName)
	sa.tokenBudget = s.tokenBudget
	return &session.CreateResponse{Session: sa}, nil
}

func (s *SessionServiceAdapter) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	sess, err := s.store.Get(req.SessionID)
	if err != nil {
		// Auto-create session if not found
		if errors.Is(err, internal.ErrSessionNotFound) {
			return s.getOrCreate(ctx, req)
		}
		return nil, err
	}
	if sess == nil {
		return s.getOrCreate(ctx, req)
	}
	sa := NewSessionAdapter(sess, s.store, s.rootAgentName)
	sa.tokenBudget = s.tokenBudget
	return &session.GetResponse{Session: sa}, nil
}

// getOrCreate attempts to create a session, and if it fails due to a
// concurrent creation (UNIQUE constraint), retries the Get instead.
func (s *SessionServiceAdapter) getOrCreate(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	createReq := &session.CreateRequest{SessionID: req.SessionID}
	resp, createErr := s.Create(ctx, createReq)
	if createErr != nil {
		// Another goroutine already created this session — fetch it.
		if errors.Is(createErr, internal.ErrDuplicateSession) {
			sess, err := s.store.Get(req.SessionID)
			if err != nil {
				return nil, fmt.Errorf("auto-create session %s: get after conflict: %w", req.SessionID, err)
			}
			sa := NewSessionAdapter(sess, s.store, s.rootAgentName)
			sa.tokenBudget = s.tokenBudget
			return &session.GetResponse{Session: sa}, nil
		}
		return nil, fmt.Errorf("auto-create session %s: %w", req.SessionID, createErr)
	}
	return &session.GetResponse{Session: resp.Session}, nil
}

func (s *SessionServiceAdapter) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	// Internal store interface doesn't strictly support List with these filters
	// We might need to extend store or minimal impl.
	// For migration, List might not be critical if Runner only uses Get/Create/AppendEvent for standard flow.
	// But let's return empty for now.
	return &session.ListResponse{}, nil
}

func (s *SessionServiceAdapter) Delete(ctx context.Context, req *session.DeleteRequest) error {
	return s.store.Delete(req.SessionID)
}

func (s *SessionServiceAdapter) AppendEvent(ctx context.Context, sess session.Session, evt *session.Event) error {
	// Map ADK event to internal message
	msg := internal.Message{
		Timestamp: evt.Timestamp,
	}

	if evt.LLMResponse.Content != nil {
		msg.Role = types.MessageRole(evt.LLMResponse.Content.Role).Normalize()

		for _, p := range evt.LLMResponse.Content.Parts {
			if p.Text != "" {
				msg.Content += p.Text
			}
			if p.FunctionCall != nil {
				argsBytes, _ := json.Marshal(p.FunctionCall.Args)
				id := p.FunctionCall.ID
				if id == "" {
					id = "call_" + p.FunctionCall.Name
				}
				tc := internal.ToolCall{
					Name:  p.FunctionCall.Name,
					Input: string(argsBytes),
					ID:    id,
				}
				msg.ToolCalls = append(msg.ToolCalls, tc)
			}
			if p.FunctionResponse != nil {
				responseBytes, _ := json.Marshal(p.FunctionResponse.Response)
				id := p.FunctionResponse.ID
				if id == "" {
					id = "call_" + p.FunctionResponse.Name
				}
				msg.ToolCalls = append(msg.ToolCalls, internal.ToolCall{
					ID:     id,
					Name:   p.FunctionResponse.Name,
					Output: string(responseBytes),
				})
				msg.Content += string(responseBytes)
			}
		}
	} else {
		// Event might be purely internal/state update without content?
		// Ensure we don't save empty messages unless necessary.
		if len(evt.Actions.StateDelta) > 0 {
			// State update event.
			// Adapt persisted metadata.
			// Currently internal model stores state in Metadata.
			// AppendEvent is for history.
			// State updates are persistent via StateStoreAdapter.
			// So we might skip appending "message" for pure state events if Lango history doesn't support them.
			return nil
		}
	}

	if msg.Role == "" {
		msg.Role = types.RoleAssistant
		if evt.Author == "user" {
			msg.Role = types.RoleUser
		}
	}

	// Preserve the ADK author for multi-agent routing.
	msg.Author = evt.Author

	if err := s.store.AppendMessage(sess.ID(), msg); err != nil {
		return err
	}

	// Update in-memory history so subsequent reads see the new event.
	// The ADK runner reads events from the same session object after appending.
	if sa, ok := sess.(*SessionAdapter); ok {
		sa.sess.History = append(sa.sess.History, msg)
	}

	return nil
}
