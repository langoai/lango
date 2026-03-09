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

// formatPartialResponse builds a response string that includes the partial
// result recovered from a timed-out or failed agent run, along with a note
// explaining the situation.
func formatPartialResponse(partial string, agentErr *adk.AgentError) string {
	note := fmt.Sprintf("\n\n---\n⚠️ %s", agentErr.UserMessage())
	return partial + note
}
