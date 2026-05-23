package live

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
)

// Session is a Lango-side wrapper around an ADK live session.
// It captures the live request channel and event iterator so the cockpit
// (or any other consumer) can interact via a single value.
type Session struct {
	inner  agent.LiveSession
	events iter.Seq2[*session.Event, error]
	cfg    Config
	closed bool
}

// New starts a new live session for the given runner/agent/session.
// Returns a Session whose Send method delivers user input and whose Events
// channel yields model output events.
func New(
	ctx context.Context,
	r *runner.Runner,
	userID, sessionID string,
	cfg Config,
	opts ...runner.RunOption,
) (*Session, error) {
	if r == nil {
		return nil, fmt.Errorf("nil runner")
	}
	liveSess, events, err := r.RunLive(ctx, userID, sessionID, cfg.toLiveRunConfig(), opts...)
	if err != nil {
		return nil, fmt.Errorf("run live: %w", err)
	}
	return &Session{
		inner:  liveSess,
		events: events,
		cfg:    cfg,
	}, nil
}

// Send delivers a user input (text or audio blob) to the live session.
func (s *Session) Send(req agent.LiveRequest) error {
	if s == nil || s.inner == nil {
		return fmt.Errorf("nil session")
	}
	if s.closed {
		return fmt.Errorf("session closed")
	}
	return s.inner.Send(req)
}

// Events returns an iterator over events emitted by the model.
// The iterator stops when the session closes or the underlying context cancels.
func (s *Session) Events() iter.Seq2[*session.Event, error] {
	if s == nil {
		return func(yield func(*session.Event, error) bool) {}
	}
	return s.events
}

// Close shuts down the live session and releases audio resources.
func (s *Session) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var innerErr error
	if s.inner != nil {
		innerErr = s.inner.Close()
	}
	if s.cfg.AudioSink != nil {
		_ = s.cfg.AudioSink.Close()
	}
	if s.cfg.AudioSource != nil {
		_ = s.cfg.AudioSource.Close()
	}
	if innerErr != nil {
		return fmt.Errorf("close live session: %w", innerErr)
	}
	return nil
}
