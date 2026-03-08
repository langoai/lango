package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/langoai/lango/internal/adk"
	"github.com/stretchr/testify/assert"
)

func TestFormatUserError_AgentError(t *testing.T) {
	t.Parallel()

	err := &adk.AgentError{
		Code:    adk.ErrTimeout,
		Message: "agent error",
		Elapsed: 30 * time.Second,
	}
	msg := FormatUserError(err)
	assert.Contains(t, msg, "[E001]")
	assert.Contains(t, msg, "timed out")
}

func TestFormatUserError_PlainError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("something went wrong")
	msg := FormatUserError(err)
	assert.Contains(t, msg, "something went wrong")
}

func TestFormatPartialResponse(t *testing.T) {
	t.Parallel()

	agentErr := &adk.AgentError{
		Code:    adk.ErrTimeout,
		Message: "timed out",
		Partial: "Here is a partial answer about...",
		Elapsed: 2 * time.Minute,
	}

	result := formatPartialResponse(agentErr.Partial, agentErr)
	assert.Contains(t, result, "Here is a partial answer about...")
	assert.Contains(t, result, "⚠️")
	assert.Contains(t, result, "[E001]")
}
