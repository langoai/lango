package chat

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

type wave52SessionStore struct {
	submitTestSessionStore

	getSession *session.Session
	getErr     error
	createErr  error
	updateErr  error

	created []*session.Session
	updated []*session.Session
}

func (s *wave52SessionStore) Get(string) (*session.Session, error) {
	return s.getSession, s.getErr
}

func (s *wave52SessionStore) Create(sess *session.Session) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.created = append(s.created, sess)
	return nil
}

func (s *wave52SessionStore) Update(sess *session.Session) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updated = append(s.updated, sess)
	return nil
}

func TestWave52CurrentModeNameHandlesStoreFailureBranches(t *testing.T) {
	m := newTestModel()

	m.sessionStore = &wave52SessionStore{getErr: errors.New("store unavailable")}
	if got := m.currentModeName(); got != "" {
		t.Fatalf("expected empty mode when store returns error, got %q", got)
	}

	m.sessionStore = &wave52SessionStore{}
	if got := m.currentModeName(); got != "" {
		t.Fatalf("expected empty mode when session is missing, got %q", got)
	}
}

func TestWave52SetSessionModePropagatesCreateAndUpdateFailures(t *testing.T) {
	createErr := errors.New("create denied")
	createStore := &wave52SessionStore{createErr: createErr}
	m := newTestModel()
	m.sessionStore = createStore

	err := m.setSessionMode("review")
	if err == nil {
		t.Fatal("expected create failure")
	}
	if !strings.Contains(err.Error(), "create session") {
		t.Fatalf("expected wrapped create error, got %v", err)
	}
	if len(createStore.updated) != 0 {
		t.Fatalf("expected no update after create failure, got %d updates", len(createStore.updated))
	}

	updateErr := errors.New("update denied")
	updateStore := &wave52SessionStore{
		getSession: &session.Session{Key: "test-session"},
		updateErr:  updateErr,
	}
	m = newTestModel()
	m.sessionStore = updateStore

	err = m.setSessionMode("review")
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error %v, got %v", updateErr, err)
	}
	if len(updateStore.created) != 0 {
		t.Fatalf("expected no create for existing session, got %d creates", len(updateStore.created))
	}
}

func TestWave52HandleStreamingEnterWithBlankDraftDoesNotCancel(t *testing.T) {
	m := newTestModel()
	m.state = stateStreaming
	m.SetComposerValue(" \n\t ")
	cancelled := false
	m.cancelFn = func() { cancelled = true }

	cmd := m.handleStreamingKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Fatalf("expected no command for blank streaming submission, got %T", cmd)
	}
	if cancelled {
		t.Fatal("blank streaming submission must not cancel the active turn")
	}
	if m.pendingRedirectInput != "" {
		t.Fatalf("expected no redirect input, got %q", m.pendingRedirectInput)
	}
	if got := m.ComposerValue(); strings.TrimSpace(got) != "" || got == "" {
		t.Fatalf("expected blank draft content to remain present, got %q", got)
	}
}

func TestWave52SubmitComposerWithParentUsesProvidedContext(t *testing.T) {
	executor := &submitCaptureExecutor{}
	m := New(Deps{
		TurnRunner:   turnrunner.New(turnrunner.Config{}, executor, submitTestSessionStore{}, nil),
		Config:       readyRemoteConfig(),
		SessionKey:   "wave52-session",
		SessionStore: submitTestSessionStore{},
	})
	parent := context.WithValue(context.Background(), wave52ContextKey{}, "parent")
	m.SetComposerValue("use parent context")

	cmd := m.SubmitComposerWithParent(parent)
	msgs := collectImmediateMsgs(cmd)

	if len(msgs) != 1 {
		t.Fatalf("expected one immediate message, got %d", len(msgs))
	}
	if _, ok := msgs[0].(DoneMsg); !ok {
		t.Fatalf("expected DoneMsg, got %T", msgs[0])
	}
	if got := executor.ctx.Value(wave52ContextKey{}); got != "parent" {
		t.Fatalf("expected runner context to inherit parent value, got %v", got)
	}
}

type wave52ContextKey struct{}
