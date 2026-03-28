package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunContext_RoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithRunContext(context.Background(), RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf-1",
		RunID:       "run-1",
	})

	rc := RunContextFromContext(ctx)
	require.NotNil(t, rc)
	assert.Equal(t, "workflow", rc.SessionType)
	assert.Equal(t, "wf-1", rc.WorkflowID)
	assert.Equal(t, "run-1", rc.RunID)
}

func TestRunContext_Absent(t *testing.T) {
	t.Parallel()

	assert.Nil(t, RunContextFromContext(context.Background()))
}

func TestTurnID_RoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithTurnID(context.Background(), "turn-abc-123")
	assert.Equal(t, "turn-abc-123", TurnIDFromContext(ctx))
}

func TestTurnID_Absent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", TurnIDFromContext(context.Background()))
}
