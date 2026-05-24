package live

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/agent"
)

func TestDefaultTextConfig_HasTextModality(t *testing.T) {
	cfg := DefaultTextConfig()
	require.Len(t, cfg.Modalities, 1)
	assert.NotNil(t, cfg.AudioSink)
	assert.NotNil(t, cfg.AudioSource)
	assert.Equal(t, 50, cfg.MaxLLMCalls)
}

func TestNew_NilRunnerErrors(t *testing.T) {
	sess, err := New(context.Background(), nil, "u", "s", DefaultTextConfig())
	require.Error(t, err)
	assert.Nil(t, sess)
}

func TestNoOpSinkAndSource_AreSafe(t *testing.T) {
	require.NoError(t, NoOpSink{}.Play(context.Background(), "audio/pcm", []byte{1, 2, 3}))
	require.NoError(t, NoOpSink{}.Close())
	require.NoError(t, NoOpSource{}.Close())
	src := NoOpSource{}
	count := 0
	for range src.Chunks(context.Background()) {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestSession_CloseIsIdempotent(t *testing.T) {
	s := &Session{}
	require.NoError(t, s.Close())
	require.NoError(t, s.Close()) // second close is a no-op
}

func TestSession_SendOnClosedReturnsErr(t *testing.T) {
	s := &Session{inner: stubLiveSession{}, closed: false}
	require.NoError(t, s.Close())
	err := s.Send(agent.LiveRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

type stubLiveSession struct{}

func (stubLiveSession) Send(agent.LiveRequest) error {
	return errors.New("stub should not be called")
}
func (stubLiveSession) Close() error { return nil }
