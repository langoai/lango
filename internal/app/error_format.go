package app

import (
	"errors"
	"fmt"

	"github.com/langoai/lango/internal/adk"
)

// FormatUserError converts an error into a user-friendly message.
// If the error is an *adk.AgentError, its structured UserMessage is used.
// Otherwise, a generic message is returned.
func FormatUserError(err error) string {
	var agentErr *adk.AgentError
	if errors.As(err, &agentErr) {
		return agentErr.UserMessage()
	}
	return fmt.Sprintf("An error occurred: %s", err.Error())
}

// formatIncompleteResponse builds a user-visible warning for a run that
// produced an internal partial draft but no safe final answer.
func formatIncompleteResponse(agentErr *adk.AgentError) string {
	return fmt.Sprintf("⚠️ %s", agentErr.UserMessage())
}
