package live

import (
	"context"
	"fmt"
	"iter"
	"sync"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// LiveExecutor opens a new live session. Local interface kept in this package
// to avoid an import cycle with internal/turnrunner (turnrunner imports live;
// live cannot import turnrunner). The default implementation lives in
// internal/turnrunner/live.go and satisfies this interface by value.
type LiveExecutor interface {
	StartLive(ctx context.Context, userID, sessionID string, cfg Config) (*Session, error)
}

// VoiceModeState describes the lifecycle of the cockpit voice mode.
type VoiceModeState int

const (
	// VoiceModeOff indicates voice mode is currently inactive.
	VoiceModeOff VoiceModeState = iota
	// VoiceModeActive indicates a live session is currently open.
	VoiceModeActive
)

// String returns a human-readable label for the state.
func (s VoiceModeState) String() string {
	switch s {
	case VoiceModeActive:
		return "active"
	default:
		return "off"
	}
}

// VoiceMode is a UI-framework-agnostic state machine that owns a Live session.
// Phase 1 supports text-only round trips; Phase 2 will populate AudioSink/Source
// in the Config passed to StartLive.
type VoiceMode struct {
	mu       sync.Mutex
	state    VoiceModeState
	executor LiveExecutor
	userID   string
	session  *Session
}

// NewVoiceMode constructs a VoiceMode bound to the given LiveExecutor and userID.
// If executor is nil, the voice mode stays permanently disabled — Toggle() errors.
func NewVoiceMode(executor LiveExecutor, userID string) *VoiceMode {
	return &VoiceMode{
		executor: executor,
		userID:   userID,
		state:    VoiceModeOff,
	}
}

// IsActive reports whether the voice mode is currently active.
func (v *VoiceMode) IsActive() bool {
	if v == nil {
		return false
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.state == VoiceModeActive
}

// State returns the current lifecycle state.
func (v *VoiceMode) State() VoiceModeState {
	if v == nil {
		return VoiceModeOff
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.state
}

// Toggle flips voice mode between Off and Active. Returns the new state
// and a non-nil error if the transition could not be completed (e.g., the
// executor is nil or starting the live session failed).
func (v *VoiceMode) Toggle(ctx context.Context, sessionID string) (VoiceModeState, error) {
	if v == nil {
		return VoiceModeOff, fmt.Errorf("nil VoiceMode")
	}
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.state == VoiceModeActive {
		if v.session != nil {
			_ = v.session.Close()
			v.session = nil
		}
		v.state = VoiceModeOff
		return VoiceModeOff, nil
	}

	if v.executor == nil {
		return VoiceModeOff, fmt.Errorf("voice mode disabled: no LiveExecutor configured")
	}

	sess, err := v.executor.StartLive(ctx, v.userID, sessionID, DefaultTextConfig())
	if err != nil {
		return VoiceModeOff, fmt.Errorf("start live: %w", err)
	}
	v.session = sess
	v.state = VoiceModeActive
	return VoiceModeActive, nil
}

// Send forwards a text message to the active live session.
func (v *VoiceMode) Send(text string) error {
	if v == nil {
		return fmt.Errorf("nil VoiceMode")
	}
	v.mu.Lock()
	sess := v.session
	v.mu.Unlock()
	if sess == nil {
		return fmt.Errorf("voice mode not active")
	}
	return sess.Send(agent.LiveRequest{
		Content: &genai.Content{
			Parts: []*genai.Part{{Text: text}},
		},
	})
}

// Events returns the live session's event stream. Returns an empty iterator
// when voice mode is not active.
func (v *VoiceMode) Events() iter.Seq2[*session.Event, error] {
	if v == nil {
		return func(yield func(*session.Event, error) bool) {}
	}
	v.mu.Lock()
	sess := v.session
	v.mu.Unlock()
	if sess == nil {
		return func(yield func(*session.Event, error) bool) {}
	}
	return sess.Events()
}

// Close ensures any active session is shut down.
func (v *VoiceMode) Close() error {
	if v == nil {
		return nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.session == nil {
		return nil
	}
	err := v.session.Close()
	v.session = nil
	v.state = VoiceModeOff
	return err
}

// Indicator returns the cockpit status string. Visibly labeled "(preview)"
// so users understand audio I/O is not yet wired.
func (v *VoiceMode) Indicator() string {
	if v == nil || !v.IsActive() {
		return ""
	}
	return "voice (preview)"
}
