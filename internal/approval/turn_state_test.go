package approval

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTurnApprovalState_RoundTrip(t *testing.T) {
	t.Parallel()

	state := NewTurnApprovalState()
	params := map[string]interface{}{"url": "https://example.com", "limit": 3}

	err := state.Put("browser_navigate", params, TurnApprovalEntry{
		Outcome:   TurnOutcomeApproved,
		Provider:  "telegram",
		RequestID: "req-1",
		Summary:   "Navigate to: https://example.com",
	})
	require.NoError(t, err)

	entry, ok, err := state.Get("browser_navigate", params)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, TurnOutcomeApproved, entry.Outcome)
	assert.Equal(t, "telegram", entry.Provider)
	assert.Equal(t, "req-1", entry.RequestID)
	assert.NotEmpty(t, entry.ParamsHash)
}

func TestTurnApprovalKey_IsDeterministic(t *testing.T) {
	t.Parallel()

	paramsA := map[string]interface{}{"url": "https://example.com", "limit": 3}
	paramsB := map[string]interface{}{"limit": 3, "url": "https://example.com"}

	keyA, hashA, err := TurnApprovalKey("browser_navigate", paramsA)
	require.NoError(t, err)

	keyB, hashB, err := TurnApprovalKey("browser_navigate", paramsB)
	require.NoError(t, err)

	assert.Equal(t, keyA, keyB)
	assert.Equal(t, hashA, hashB)
}

func TestTurnApprovalKey_BrowserSearchIgnoresLimitVariants(t *testing.T) {
	t.Parallel()

	queryOnly := map[string]interface{}{"query": "Trump   latest news"}
	limitThree := map[string]interface{}{"query": "  Trump latest   news ", "limit": 3}
	limitFive := map[string]interface{}{"query": "Trump latest news", "limit": 5}

	keyA, hashA, err := TurnApprovalKey("browser_search", queryOnly)
	require.NoError(t, err)

	keyB, hashB, err := TurnApprovalKey("browser_search", limitThree)
	require.NoError(t, err)

	keyC, hashC, err := TurnApprovalKey("browser_search", limitFive)
	require.NoError(t, err)

	assert.Equal(t, keyA, keyB)
	assert.Equal(t, keyA, keyC)
	assert.Equal(t, hashA, hashB)
	assert.Equal(t, hashA, hashC)
}

func TestWithTurnApprovalState_ContextRoundTrip(t *testing.T) {
	t.Parallel()

	state := NewTurnApprovalState()
	ctx := WithTurnApprovalState(context.Background(), state)

	assert.Same(t, state, TurnApprovalStateFromContext(ctx))
	assert.Nil(t, TurnApprovalStateFromContext(context.Background()))
}
