package voice

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// GeminiSTTConfig configures the Gemini-backed STT provider.
type GeminiSTTConfig struct {
	Client *genai.Client
	// Model is the Gemini model used for transcription.
	// Default: "gemini-2.5-flash" (handles audio inputs in v1.57+).
	Model string
}

// NewGeminiSTT constructs a Gemini-backed SpeechToText. Returns NoOpSTT
// if Client is nil — useful for graceful-degradation paths.
func NewGeminiSTT(cfg GeminiSTTConfig) SpeechToText {
	if cfg.Client == nil {
		return NoOpSTT{}
	}
	model := cfg.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &geminiSTT{client: cfg.Client, model: model}
}

type geminiSTT struct {
	client *genai.Client
	model  string
}

// Transcribe implements SpeechToText by calling GenerateContent with an
// audio Blob input and a transcription prompt. The model is asked to
// return only the transcribed text.
func (g *geminiSTT) Transcribe(ctx context.Context, audio Audio, opts STTOptions) (string, error) {
	if g == nil {
		return "", ErrNotConfigured
	}
	if len(audio.Data) == 0 {
		return "", fmt.Errorf("empty audio data")
	}
	if g.client == nil {
		return "", ErrNotConfigured
	}
	prompt := "Transcribe the audio verbatim. Output only the transcription text, no commentary."
	if opts.Language != "" {
		prompt = fmt.Sprintf("Transcribe the audio in %s, verbatim. Output only the transcription text, no commentary.", opts.Language)
	}
	resp, err := g.client.Models.GenerateContent(ctx, g.model, []*genai.Content{
		{Parts: []*genai.Part{
			{Text: prompt},
			{InlineData: &genai.Blob{MIMEType: audio.MimeType, Data: audio.Data}},
		}},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("gemini transcribe: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("gemini transcribe: nil response")
	}
	return strings.TrimSpace(resp.Text()), nil
}

// GeminiTTSConfig configures the Gemini-backed TTS provider.
type GeminiTTSConfig struct {
	Client *genai.Client
	// Model must be a Gemini TTS-capable model (e.g. "gemini-2.0-flash-exp-tts").
	// Empty disables TTS — Synthesize returns ErrTTSNotConfigured.
	Model string
}

// NewGeminiTTS constructs a Gemini-backed TextToSpeech. If Model is empty
// or Client is nil, the returned implementation always returns
// ErrTTSNotConfigured — allowing channels to skip the TTS step gracefully.
func NewGeminiTTS(cfg GeminiTTSConfig) TextToSpeech {
	if cfg.Client == nil || cfg.Model == "" {
		return NoOpTTS{}
	}
	return &geminiTTS{client: cfg.Client, model: cfg.Model}
}

type geminiTTS struct {
	client *genai.Client
	model  string
}

// Synthesize implements TextToSpeech. Phase 1 returns ErrTTSNotConfigured
// because the Gemini TTS audio-output flow requires careful per-model
// prompt construction and audio-blob extraction from the response — Phase 2
// will implement this when per-channel TTS wiring is also added.
func (g *geminiTTS) Synthesize(ctx context.Context, text string, opts TTSOptions) (Audio, error) {
	return Audio{}, fmt.Errorf("gemini tts: %w (Phase 2 will implement)", ErrTTSNotConfigured)
}

// Compile-time interface checks.
var (
	_ SpeechToText = (*geminiSTT)(nil)
	_ TextToSpeech = (*geminiTTS)(nil)
)
