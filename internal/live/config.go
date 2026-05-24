// Package live provides a thin Lango-side wrapper around ADK's Live API.
//
// Phase 1 (this package): plumbing layer — Session struct, LiveExecutor adapter,
// no-op audio interfaces. Text-only Live mode is fully functional.
//
// Phase 2 (future): malgo / portaudio bindings, PCM resampling, gemini-flash-live
// audio modality wiring. The interfaces in audio.go are designed so Phase 2 can
// plug audio in without changes to Session or its consumers.
package live

import (
	"google.golang.org/adk/agent"
	"google.golang.org/genai"
)

// Config describes how a live session should be started.
type Config struct {
	// Modalities controls which response channels the model produces.
	// Phase 1 supports text only; Phase 2 will add audio.
	Modalities []genai.Modality

	// AudioSink receives audio output from the model when an audio modality
	// is enabled. Default: NoOpSink (drops audio). Phase 2 will provide a
	// speaker-backed implementation.
	AudioSink AudioSink

	// AudioSource produces audio input to be sent to the model. Default:
	// NoOpSource (yields nothing). Phase 2 will provide a mic-backed
	// implementation.
	AudioSource AudioSource

	// MaxLLMCalls caps the number of model calls inside the live session.
	// Zero means use the upstream default.
	MaxLLMCalls int
}

// DefaultTextConfig returns a Config that runs the live session in text-only
// mode with no audio sink/source. Suitable for the Phase 1 cockpit voice mode.
func DefaultTextConfig() Config {
	return Config{
		Modalities:  []genai.Modality{genai.ModalityText},
		AudioSink:   NoOpSink{},
		AudioSource: NoOpSource{},
		MaxLLMCalls: 50,
	}
}

// toLiveRunConfig converts Lango's Config into ADK's LiveRunConfig.
func (c Config) toLiveRunConfig() agent.LiveRunConfig {
	return agent.LiveRunConfig{
		ResponseModalities: c.Modalities,
		MaxLLMCalls:        c.MaxLLMCalls,
		SaveLiveBlob:       false, // privacy default; Phase 2 may opt in
	}
}
