package voice

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpSTT_TranscribeReturnsErrNotConfigured(t *testing.T) {
	out, err := NoOpSTT{}.Transcribe(context.Background(), Audio{}, STTOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotConfigured)
	assert.Empty(t, out)
}

func TestNoOpTTS_SynthesizeReturnsErrTTSNotConfigured(t *testing.T) {
	out, err := NoOpTTS{}.Synthesize(context.Background(), "hello", TTSOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTTSNotConfigured)
	assert.Empty(t, out.Data)
}

func TestNewGeminiSTT_NilClientReturnsNoOp(t *testing.T) {
	stt := NewGeminiSTT(GeminiSTTConfig{Client: nil})
	require.NotNil(t, stt)
	_, err := stt.Transcribe(context.Background(), Audio{Data: []byte{1}}, STTOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotConfigured)
}

func TestNewGeminiTTS_EmptyModelReturnsNoOp(t *testing.T) {
	tts := NewGeminiTTS(GeminiTTSConfig{Client: nil, Model: ""})
	require.NotNil(t, tts)
	_, err := tts.Synthesize(context.Background(), "x", TTSOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTTSNotConfigured)
}

func TestGeminiSTT_EmptyAudioErrors(t *testing.T) {
	// Construct a geminiSTT directly to test the early validation
	// without making a network call.
	g := &geminiSTT{model: "gemini-2.5-flash"}
	_, err := g.Transcribe(context.Background(), Audio{}, STTOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty audio data")
}

func TestGeminiSTT_NilReceiverReturnsErrNotConfigured(t *testing.T) {
	var g *geminiSTT
	_, err := g.Transcribe(context.Background(), Audio{Data: []byte{1}}, STTOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotConfigured)
}

func TestGeminiTTS_AlwaysReturnsErrInPhase1(t *testing.T) {
	g := &geminiTTS{model: "gemini-2.0-flash-exp-tts"}
	_, err := g.Synthesize(context.Background(), "hi", TTSOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTTSNotConfigured), "expected ErrTTSNotConfigured wrapped")
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	require.NotEqual(t, ErrNotConfigured, ErrTTSNotConfigured)
	require.False(t, errors.Is(ErrNotConfigured, ErrTTSNotConfigured))
	require.False(t, errors.Is(ErrTTSNotConfigured, ErrNotConfigured))
}
