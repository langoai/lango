package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/session"
)

// TextGenerator generates text from a prompt. Used to abstract LLM calls for testability.
type TextGenerator interface {
	GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

const observerPrompt = `You are a conversation memory assistant. Your task is to compress the following conversation messages into a concise observation note.

Focus on capturing:
- Key decisions made
- User's intent and goals
- Important facts and context mentioned
- Task progress and outcomes
- Any action items or next steps

Do NOT include:
- Verbatim tool outputs or code blocks
- Redundant greetings or pleasantries
- Technical details that can be re-derived

Write a concise paragraph (2-5 sentences) capturing the essential information.`

// Observer generates compressed observation notes from conversation history.
type Observer struct {
	generator TextGenerator
	store     *Store
	logger    *zap.SugaredLogger
}

// NewObserver creates a new Observer.
func NewObserver(generator TextGenerator, store *Store, logger *zap.SugaredLogger) *Observer {
	return &Observer{
		generator: generator,
		store:     store,
		logger:    logger,
	}
}

// Observe generates a compressed observation from un-observed messages.
// It takes messages from lastObservedIndex+1 onward, calls the LLM, and persists the result.
// Returns nil if there are no new messages to observe.
func (o *Observer) Observe(ctx context.Context, sessionKey string, messages []session.Message, lastObservedIndex int) (*Observation, error) {
	startIdx := lastObservedIndex + 1
	if startIdx >= len(messages) {
		return nil, nil
	}

	newMessages := messages[startIdx:]
	if len(newMessages) == 0 {
		return nil, nil
	}

	userPrompt := formatMessages(newMessages)

	response, err := o.generator.GenerateText(ctx, observerPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("generate observation: %w", err)
	}

	obs := Observation{
		ID:               uuid.New(),
		SessionKey:       sessionKey,
		Content:          response,
		TokenCount:       EstimateTokens(response),
		SourceStartIndex: startIdx,
		SourceEndIndex:   len(messages) - 1,
		CreatedAt:        time.Now(),
	}

	if err := o.store.SaveObservation(ctx, obs); err != nil {
		return nil, fmt.Errorf("save observation: %w", err)
	}

	o.logger.Debugw("observation created",
		"sessionKey", sessionKey,
		"sourceRange", fmt.Sprintf("%d-%d", obs.SourceStartIndex, obs.SourceEndIndex),
		"tokens", obs.TokenCount,
	)

	return &obs, nil
}

// formatMessages formats messages into a text block for the LLM prompt.
func formatMessages(msgs []session.Message) string {
	var b strings.Builder
	for _, msg := range msgs {
		fmt.Fprintf(&b, "[%s]: %s\n", msg.Role, msg.Content)
		for _, tc := range msg.ToolCalls {
			fmt.Fprintf(&b, "  [tool:%s] %s\n", tc.Name, tc.Input)
			if tc.Output != "" {
				fmt.Fprintf(&b, "  [result] %s\n", tc.Output)
			}
		}
	}
	return b.String()
}
