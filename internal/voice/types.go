// Package voice provides STT (speech-to-text) and TTS (text-to-speech)
// abstractions for Lango. Phase 1 ships the interfaces, a Gemini-backed
// provider, and no-op defaults; per-channel integration is Phase 2.
package voice

import "errors"

// ErrNotConfigured is returned when a voice provider is invoked but no
// configuration has been supplied. Channels should detect this and fall
// back to text-only behavior.
var ErrNotConfigured = errors.New("voice not configured")

// ErrTTSNotConfigured is returned specifically when TTS is unavailable.
// Useful when STT is configured but TTS is not (asymmetric setup).
var ErrTTSNotConfigured = errors.New("text-to-speech not configured")

// Audio is a single audio chunk with its MIME type.
// Common MIME types:
//
//	audio/ogg;codecs=opus    — Telegram voice messages
//	audio/mp3                — Slack file uploads
//	audio/pcm;rate=16000     — Gemini Live input
//	audio/pcm;rate=24000     — Gemini Live output
type Audio struct {
	MimeType string
	Data     []byte
}

// STTOptions configures a single STT invocation.
type STTOptions struct {
	// Language is a BCP-47 tag, e.g. "ko-KR" or "en-US". Empty means auto-detect.
	Language string
}

// TTSOptions configures a single TTS invocation.
type TTSOptions struct {
	// Language is a BCP-47 tag, e.g. "ko-KR" or "en-US".
	Language string

	// Voice is the named voice identifier (provider-specific). Empty means default.
	Voice string
}
