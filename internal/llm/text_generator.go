// Package llm provides common LLM abstraction interfaces.
package llm

import "context"

// TextGenerator generates text from a system prompt and user prompt pair.
// It abstracts LLM calls so that callers remain provider-agnostic and testable.
type TextGenerator interface {
	GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
