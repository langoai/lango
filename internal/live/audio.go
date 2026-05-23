package live

import (
	"context"
	"iter"
)

// AudioSink consumes audio chunks emitted by the model in a live session.
type AudioSink interface {
	// Play accepts a chunk of PCM (or codec-specific) audio bytes for playback.
	// Implementations should be non-blocking when possible — block only when
	// the playback queue is full.
	Play(ctx context.Context, mimeType string, audio []byte) error

	// Close releases any audio resources.
	Close() error
}

// AudioSource produces audio chunks to send to the model in a live session.
type AudioSource interface {
	// Chunks returns an iterator over captured audio chunks. The iterator
	// stops when the source is closed or context is cancelled.
	Chunks(ctx context.Context) iter.Seq2[AudioChunk, error]

	// Close releases the underlying capture device.
	Close() error
}

// AudioChunk is a single captured audio packet.
type AudioChunk struct {
	MimeType string // e.g. "audio/pcm;rate=16000"
	Data     []byte
}

// NoOpSink is the default AudioSink — drops all audio.
// Phase 2 will provide a malgo-backed speaker sink.
type NoOpSink struct{}

// Play implements AudioSink by discarding the supplied audio.
func (NoOpSink) Play(context.Context, string, []byte) error { return nil }

// Close implements AudioSink.
func (NoOpSink) Close() error { return nil }

// NoOpSource is the default AudioSource — yields nothing.
// Phase 2 will provide a malgo-backed mic source.
type NoOpSource struct{}

// Chunks implements AudioSource by returning an immediately-completing iterator.
func (NoOpSource) Chunks(context.Context) iter.Seq2[AudioChunk, error] {
	return func(yield func(AudioChunk, error) bool) {}
}

// Close implements AudioSource.
func (NoOpSource) Close() error { return nil }
