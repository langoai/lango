package adk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/session"
)

type SessionServiceAdapter struct {
	store               internal.Store
	rootAgentName       string
	tokenBudget         int // 0 = use DefaultTokenBudget
	rootSessionObserver func(string)
	childStore          *internal.InMemoryChildStore
	summarizer          Summarizer
	isolatedAgents      map[string]bool
	childMu             sync.Mutex
	activeChild         map[string]*runtimeChild
}

type runtimeChild struct {
	key      string
	agent    string
	child    *internal.ChildSession
	parentID string
}

func NewSessionServiceAdapter(store internal.Store, rootAgentName string) *SessionServiceAdapter {
	return &SessionServiceAdapter{
		store:          store,
		rootAgentName:  rootAgentName,
		activeChild:    make(map[string]*runtimeChild),
		summarizer:     &StructuredSummarizer{},
		isolatedAgents: make(map[string]bool),
	}
}

// WithTokenBudget sets the token budget for history truncation.
// Use ModelTokenBudget(modelName) to derive an appropriate budget from the model name.
func (s *SessionServiceAdapter) WithTokenBudget(budget int) *SessionServiceAdapter {
	s.tokenBudget = budget
	return s
}

// WithRootSessionObserver records root session creation events.
func (s *SessionServiceAdapter) WithRootSessionObserver(fn func(string)) *SessionServiceAdapter {
	s.rootSessionObserver = fn
	return s
}

// WithChildLifecycleHook enables synthetic child-session lifecycle tracking.
func (s *SessionServiceAdapter) WithChildLifecycleHook(h func(internal.SessionLifecycleEvent)) *SessionServiceAdapter {
	if h != nil {
		s.childStore = internal.NewInMemoryChildStore(s.store, internal.WithLifecycleHook(h))
	}
	return s
}

// WithIsolatedAgents marks the agent names that should write to child session history.
func (s *SessionServiceAdapter) WithIsolatedAgents(names []string) *SessionServiceAdapter {
	s.isolatedAgents = make(map[string]bool, len(names))
	for _, name := range names {
		if strings.TrimSpace(name) != "" {
			s.isolatedAgents[name] = true
		}
	}
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
	if s.rootSessionObserver != nil {
		s.rootSessionObserver(req.SessionID)
	}

	sa := NewSessionAdapter(sess, s.store, s.rootAgentName)
	sa.tokenBudget = s.tokenBudget
	return &session.CreateResponse{Session: sa}, nil
}

// Get retrieves a session by ID.
//
// CONTRACT DEVIATION: ADK's session.Service.Get() contract expects an error for
// missing sessions. This implementation auto-creates missing sessions and
// auto-renews expired sessions instead of returning an error. This is intentional
// because lango's session lifecycle is self-managing — the caller should not need
// to handle "not found" as a special case. The auto-create/renew behavior is
// preserved for backward compatibility and must not be changed without updating
// all callers.
func (s *SessionServiceAdapter) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	sess, err := s.store.Get(req.SessionID)
	if err != nil {
		// Auto-create session if not found
		if errors.Is(err, internal.ErrSessionNotFound) {
			return s.getOrCreate(ctx, req)
		}
		// Auto-renew expired sessions: delete stale record, then create fresh
		if errors.Is(err, internal.ErrSessionExpired) {
			logger().Infow("session expired, auto-renewing", "session", req.SessionID)
			if delErr := s.store.Delete(req.SessionID); delErr != nil {
				return nil, fmt.Errorf("delete expired session %s: %w", req.SessionID, delErr)
			}
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
	s.trackChildLifecycle(evt, sess.ID())

	// Map ADK event to internal message
	msg, skip, err := eventToMessage(evt)
	if err != nil {
		return err
	}
	if skip {
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
	return s.appendMessage(sess, msg)
}

// CloseActiveChild merges any active synthetic child session for the parent session.
func (s *SessionServiceAdapter) CloseActiveChild(sessionID string) error {
	if s.childStore == nil {
		return nil
	}
	s.childMu.Lock()
	active := s.activeChild[sessionID]
	if active == nil {
		s.childMu.Unlock()
		return nil
	}
	delete(s.activeChild, sessionID)
	s.childMu.Unlock()

	summary := ""
	if s.summarizer != nil {
		var err error
		summary, err = s.summarizer.Summarize(active.child.History)
		if err != nil {
			return err
		}
	}
	return s.childStore.MergeChildAsAuthor(active.key, summary, s.rootAgentName)
}

// DiscardActiveChild discards the current synthetic child session for the parent session.
func (s *SessionServiceAdapter) DiscardActiveChild(sessionID string) error {
	if s.childStore == nil {
		return nil
	}
	s.childMu.Lock()
	active := s.activeChild[sessionID]
	if active == nil {
		s.childMu.Unlock()
		return nil
	}
	delete(s.activeChild, sessionID)
	s.childMu.Unlock()
	return s.childStore.DiscardChild(active.key)
}

func (s *SessionServiceAdapter) trackChildLifecycle(evt *session.Event, sessionID string) {
	if s.childStore == nil || evt == nil {
		return
	}

	author := evt.Author
	if author == "" || author == "user" || author == s.rootAgentName || !s.isolatedAgents[author] {
		_ = s.CloseActiveChild(sessionID)
		return
	}

	s.childMu.Lock()
	active := s.activeChild[sessionID]
	s.childMu.Unlock()
	if active != nil && active.agent == author {
		if evt.Actions.TransferToAgent == s.rootAgentName && !hasText(evt) && !hasFunctionCalls(evt) {
			_ = s.DiscardActiveChild(sessionID)
		}
		return
	}

	_ = s.CloseActiveChild(sessionID)
	s.forkChildSession(evt, sessionID, author)
}

// forkChildSession creates a new synthetic child session for the given author
// and registers it as the active child. If the event immediately transfers back
// to the root agent without content, the child is discarded.
func (s *SessionServiceAdapter) forkChildSession(evt *session.Event, sessionID, author string) {
	child, err := s.childStore.ForkChild(sessionID, author, internal.ChildSessionConfig{
		SummarizeOnMerge: true,
	})
	if err != nil {
		logger().Debugw("fork synthetic child session", "session", sessionID, "author", author, "error", err)
		return
	}

	s.childMu.Lock()
	s.activeChild[sessionID] = &runtimeChild{key: child.Key, agent: author, child: child, parentID: sessionID}
	s.childMu.Unlock()

	if evt.Actions.TransferToAgent == s.rootAgentName && !hasText(evt) && !hasFunctionCalls(evt) {
		_ = s.DiscardActiveChild(sessionID)
	}
}

func (s *SessionServiceAdapter) appendMessage(sess session.Session, msg internal.Message) error {
	targetID := sess.ID()
	var parentAdapter *SessionAdapter
	if sa, ok := sess.(*SessionAdapter); ok {
		parentAdapter = sa
	}

	s.childMu.Lock()
	active := s.activeChild[targetID]
	s.childMu.Unlock()

	if active != nil && s.isolatedAgents[msg.Author] {
		active.child.AppendMessage(msg)
		return nil
	}

	if err := s.store.AppendMessage(targetID, msg); err != nil {
		return err
	}
	if parentAdapter != nil {
		parentAdapter.sess.History = append(parentAdapter.sess.History, msg)
	}
	return nil
}

func eventToMessage(evt *session.Event) (internal.Message, bool, error) {
	msg := internal.Message{
		Timestamp: evt.Timestamp,
	}

	content := evt.Content
	if content == nil && evt.LLMResponse.Content != nil {
		content = evt.LLMResponse.Content
	}

	if content != nil {
		msg.Role = types.MessageRole(content.Role).Normalize()

		for _, p := range content.Parts {
			if p.Text != "" {
				msg.Content += p.Text
			}
			if p.FunctionCall != nil {
				msg.ToolCalls = append(msg.ToolCalls, functionCallToToolCall(p.FunctionCall, p))
			}
			if p.FunctionResponse != nil {
				tc := functionResponseToToolCall(p.FunctionResponse)
				msg.ToolCalls = append(msg.ToolCalls, tc)
				msg.Content += tc.Output
			}
		}
		hasFuncResponse := false
		hasFuncCall := false
		for _, tc := range msg.ToolCalls {
			if tc.Output != "" {
				hasFuncResponse = true
			}
			if tc.Input != "" {
				hasFuncCall = true
			}
		}
		if hasFuncResponse && !hasFuncCall {
			msg.Role = types.RoleTool
		}
	} else if len(evt.Actions.StateDelta) > 0 {
		return msg, true, nil
	}

	if msg.Role == "" {
		msg.Role = types.RoleAssistant
		if evt.Author == "user" {
			msg.Role = types.RoleUser
		}
	}
	msg.Author = evt.Author
	return msg, false, nil
}
