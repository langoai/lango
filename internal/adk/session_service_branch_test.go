package adk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/session"
)

type sessionServiceBranchStore struct {
	*mockStore
	createErr error
	getErrs   []error

	createCalls int
	getCalls    int
}

func (s *sessionServiceBranchStore) Create(sess *internal.Session) error {
	s.createCalls++
	if s.createErr != nil {
		return s.createErr
	}
	return s.mockStore.Create(sess)
}

func (s *sessionServiceBranchStore) Get(key string) (*internal.Session, error) {
	s.getCalls++
	if s.getCalls <= len(s.getErrs) && s.getErrs[s.getCalls-1] != nil {
		return nil, s.getErrs[s.getCalls-1]
	}
	return s.mockStore.Get(key)
}

type sessionServiceBranchSummarizer struct {
	summary string
	err     error
}

func (s sessionServiceBranchSummarizer) Summarize([]internal.Message) (string, error) {
	return s.summary, s.err
}

func TestSessionServiceAdapterGetHandlesAutoCreateConflictAndErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("duplicate create fetches existing session", func(t *testing.T) {
		t.Parallel()

		store := &sessionServiceBranchStore{
			mockStore: newMockStore(),
			createErr: internal.ErrDuplicateSession,
			getErrs:   []error{internal.ErrSessionNotFound},
		}
		store.sessions["race-session"] = &internal.Session{
			Key:       "race-session",
			Metadata:  make(map[string]string),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		svc := NewSessionServiceAdapter(store, "lango-agent").WithTokenBudget(123)

		resp, err := svc.Get(ctx, &session.GetRequest{SessionID: "race-session"})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "race-session", resp.Session.ID())
		adapter, ok := resp.Session.(*SessionAdapter)
		require.True(t, ok)
		assert.Equal(t, 123, adapter.tokenBudget)
		assert.Equal(t, 1, store.createCalls)
		assert.Equal(t, 2, store.getCalls)
	})

	t.Run("duplicate create wraps get after conflict failure", func(t *testing.T) {
		t.Parallel()

		getErr := errors.New("store unavailable")
		store := &sessionServiceBranchStore{
			mockStore: newMockStore(),
			createErr: internal.ErrDuplicateSession,
			getErrs:   []error{internal.ErrSessionNotFound, getErr},
		}
		svc := NewSessionServiceAdapter(store, "lango-agent")

		resp, err := svc.Get(ctx, &session.GetRequest{SessionID: "race-session"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, getErr)
		assert.Contains(t, err.Error(), "get after conflict")
	})

	t.Run("non duplicate create failure is wrapped", func(t *testing.T) {
		t.Parallel()

		createErr := errors.New("read-only store")
		store := &sessionServiceBranchStore{
			mockStore: newMockStore(),
			createErr: createErr,
			getErrs:   []error{internal.ErrSessionNotFound},
		}
		svc := NewSessionServiceAdapter(store, "lango-agent")

		resp, err := svc.Get(ctx, &session.GetRequest{SessionID: "new-session"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, createErr)
		assert.Contains(t, err.Error(), "auto-create session new-session")
		assert.Equal(t, 1, store.getCalls)
	})
}

func TestCloseActiveChildHandlesSummaryFallbackAndErrors(t *testing.T) {
	t.Parallel()

	t.Run("summarizer error rolls back overlay without merging", func(t *testing.T) {
		t.Parallel()

		store := newMockStore()
		sess := &internal.Session{
			Key:       "test-session",
			Metadata:  make(map[string]string),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		store.Create(sess)
		adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
		svc := NewSessionServiceAdapter(store, "lango-orchestrator").
			WithIsolatedAgents([]string{"operator"})
		wantErr := errors.New("summary backend failed")
		svc.summarizer = sessionServiceBranchSummarizer{err: wantErr}

		require.NoError(t, svc.AppendEvent(context.Background(), adapter, newTestEvent("operator", "model", "private result")))

		err := svc.CloseActiveChild("test-session")

		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)
		assert.Empty(t, adapter.sess.History, "parent-visible overlay should be rolled back before returning")
		assert.Empty(t, store.messages["test-session"], "failed summaries must not merge child history")
		assert.Nil(t, svc.activeChild["test-session"], "failed close still consumes the active child")
	})

	t.Run("blank summary falls back to explicit empty result note", func(t *testing.T) {
		t.Parallel()

		store := newMockStore()
		sess := &internal.Session{
			Key:       "test-session",
			Metadata:  make(map[string]string),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		store.Create(sess)
		adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
		svc := NewSessionServiceAdapter(store, "lango-orchestrator").
			WithIsolatedAgents([]string{"operator"})
		svc.summarizer = sessionServiceBranchSummarizer{summary: "   "}

		require.NoError(t, svc.AppendEvent(context.Background(), adapter, newTestEvent("operator", "model", "private result")))
		require.NoError(t, svc.CloseActiveChild("test-session"))

		dbMsgs := store.messages["test-session"]
		require.Len(t, dbMsgs, 1)
		assert.Equal(t, "lango-orchestrator", dbMsgs[0].Author)
		assert.Equal(t, "[Isolated sub-agent operator ended without a visible assistant result: empty_after_tool_use.]", dbMsgs[0].Content)
		require.Len(t, adapter.sess.History, 1)
		assert.Equal(t, dbMsgs[0].Content, adapter.sess.History[0].Content)
	})
}

func TestCleanupFailedTurnReturnsDiscardErrorAfterRollback(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		History: []internal.Message{
			{
				Role:      types.RoleUser,
				Content:   "root request",
				Timestamp: time.Now(),
				Author:    "user",
			},
			{
				Role:      types.RoleAssistant,
				Content:   "private overlay",
				Timestamp: time.Now(),
				Author:    "operator",
			},
		},
	}
	store.Create(sess)
	adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
	svc := NewSessionServiceAdapter(store, "lango-orchestrator")
	svc.childStore = internal.NewInMemoryChildStore(store)
	svc.activeChild["test-session"] = &runtimeChild{
		key:         "missing-child",
		agent:       "operator",
		child:       &internal.ChildSession{Key: "missing-child", ParentKey: "test-session", AgentName: "operator"},
		parentID:    "test-session",
		parent:      adapter,
		baseHistory: 1,
		overlayLen:  1,
	}

	err := svc.CleanupFailedTurn("test-session", "agent error")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `child session "missing-child" not found`)
	require.Len(t, adapter.sess.History, 1)
	assert.Equal(t, "root request", adapter.sess.History[0].Content)
	assert.Empty(t, store.messages["test-session"], "failure note should not be appended after discard failure")
	assert.Nil(t, svc.activeChild["test-session"])
}
