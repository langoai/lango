package turntrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventTypeConstants_AreUnique(t *testing.T) {
	constants := []EventType{
		EventToolCall, EventToolResult, EventDelegation, EventDelegationReturn,
		EventText, EventTerminalError, EventBudgetWarning, EventRecoveryAttempt,
		EventPolicyDecision,
	}
	seen := make(map[EventType]struct{}, len(constants))
	for _, c := range constants {
		assert.NotEmpty(t, c, "event type constant must not be empty")
		_, dup := seen[c]
		assert.False(t, dup, "duplicate event type constant: %q", c)
		seen[c] = struct{}{}
	}
}

func TestEventType_IsStringAlias(t *testing.T) {
	// EventType is a type alias (not a new type), so it must be assignable to string.
	var s = EventToolCall
	assert.Equal(t, "tool_call", s)
}
