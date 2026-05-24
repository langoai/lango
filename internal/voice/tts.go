package voice

import "context"

// TextToSpeech synthesizes audio from text.
type TextToSpeech interface {
	// Synthesize converts text into an audio chunk. The returned MimeType
	// indicates the output codec. Implementations should return
	// ErrTTSNotConfigured when TTS is unavailable.
	Synthesize(ctx context.Context, text string, opts TTSOptions) (Audio, error)
}
