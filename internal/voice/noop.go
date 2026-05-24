package voice

import "context"

// NoOpSTT always returns ErrNotConfigured. Use it when STT is disabled
// so calling code can branch on the error without crashing.
type NoOpSTT struct{}

// Transcribe implements SpeechToText.
func (NoOpSTT) Transcribe(context.Context, Audio, STTOptions) (string, error) {
	return "", ErrNotConfigured
}

// NoOpTTS always returns ErrTTSNotConfigured.
type NoOpTTS struct{}

// Synthesize implements TextToSpeech.
func (NoOpTTS) Synthesize(context.Context, string, TTSOptions) (Audio, error) {
	return Audio{}, ErrTTSNotConfigured
}

// Compile-time interface checks.
var (
	_ SpeechToText = NoOpSTT{}
	_ TextToSpeech = NoOpTTS{}
)
