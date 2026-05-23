package voice

import "context"

// SpeechToText transcribes audio into text.
type SpeechToText interface {
	// Transcribe converts an audio chunk into a textual transcription.
	// Implementations should return ErrNotConfigured when the provider
	// is invoked without proper setup.
	Transcribe(ctx context.Context, audio Audio, opts STTOptions) (string, error)
}
