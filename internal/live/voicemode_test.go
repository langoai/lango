package live

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubLiveExecutor struct {
	err     error
	started int
}

func (s *stubLiveExecutor) StartLive(ctx context.Context, userID, sessionID string, cfg Config) (*Session, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.started++
	// Return a zero-value Session — Close is nil-safe; tests do not exercise Send/Events.
	return &Session{}, nil
}

func TestNewVoiceMode_StartsOff(t *testing.T) {
	v := NewVoiceMode(nil, "u")
	assert.False(t, v.IsActive())
	assert.Equal(t, VoiceModeOff, v.State())
	assert.Equal(t, "", v.Indicator())
}

func TestVoiceMode_ToggleWithoutExecutorErrors(t *testing.T) {
	v := NewVoiceMode(nil, "u")
	state, err := v.Toggle(context.Background(), "s")
	require.Error(t, err)
	assert.Equal(t, VoiceModeOff, state)
	assert.Contains(t, err.Error(), "disabled")
}

func TestVoiceMode_ToggleStartsAndStops(t *testing.T) {
	exec := &stubLiveExecutor{}
	v := NewVoiceMode(exec, "u")

	state, err := v.Toggle(context.Background(), "s")
	require.NoError(t, err)
	assert.Equal(t, VoiceModeActive, state)
	assert.True(t, v.IsActive())
	assert.Equal(t, "voice (preview)", v.Indicator())
	assert.Equal(t, 1, exec.started)

	state, err = v.Toggle(context.Background(), "s")
	require.NoError(t, err)
	assert.Equal(t, VoiceModeOff, state)
	assert.False(t, v.IsActive())
}

func TestVoiceMode_ToggleStartFailureKeepsOff(t *testing.T) {
	exec := &stubLiveExecutor{err: errors.New("boom")}
	v := NewVoiceMode(exec, "u")
	state, err := v.Toggle(context.Background(), "s")
	require.Error(t, err)
	assert.Equal(t, VoiceModeOff, state)
	assert.False(t, v.IsActive())
}

func TestVoiceMode_SendWhenInactive(t *testing.T) {
	v := NewVoiceMode(nil, "u")
	err := v.Send("hi")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestVoiceMode_EventsWhenInactiveReturnsEmpty(t *testing.T) {
	v := NewVoiceMode(nil, "u")
	count := 0
	for range v.Events() {
		count++
	}
	assert.Equal(t, 0, count)
}
