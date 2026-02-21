package memory

import (
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/types"
)

// perMessageOverhead is the token overhead per message for role/formatting.
const perMessageOverhead = 4

// EstimateTokens delegates to types.EstimateTokens for token count approximation.
func EstimateTokens(text string) int {
	return types.EstimateTokens(text)
}

// CountMessageTokens returns the estimated token count for a single message.
func CountMessageTokens(msg session.Message) int {
	tokens := perMessageOverhead
	tokens += types.EstimateTokens(msg.Content)
	for _, tc := range msg.ToolCalls {
		tokens += types.EstimateTokens(tc.ID)
		tokens += types.EstimateTokens(tc.Name)
		tokens += types.EstimateTokens(tc.Input)
		tokens += types.EstimateTokens(tc.Output)
	}
	return tokens
}

// CountMessagesTokens returns the total estimated token count for a batch of messages.
func CountMessagesTokens(msgs []session.Message) int {
	var total int
	for _, msg := range msgs {
		total += CountMessageTokens(msg)
	}
	return total
}
