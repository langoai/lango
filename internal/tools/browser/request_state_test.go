package browser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestState_RecordSearchWarnsAtThirdCall(t *testing.T) {
	t.Parallel()

	state := NewRequestState()

	count, queries, warn, limited := state.RecordSearch("first query", "https://example.com/1")
	assert.Equal(t, 1, count)
	assert.Equal(t, []string{"first query"}, queries)
	assert.False(t, warn)
	assert.False(t, limited)

	count, queries, warn, limited = state.RecordSearch("second query", "https://example.com/2")
	assert.Equal(t, 2, count)
	assert.Equal(t, []string{"first query", "second query"}, queries)
	assert.False(t, warn)
	assert.False(t, limited)

	count, queries, warn, limited = state.RecordSearch("third query", "https://example.com/3")
	assert.Equal(t, 3, count)
	assert.Equal(t, []string{"first query", "second query", "third query"}, queries)
	assert.True(t, warn)
	assert.True(t, limited)
	assert.Equal(t, "https://example.com/3", state.CurrentURL())

	count, _, warn, limited = state.RecordSearch("fourth query", "https://example.com/4")
	assert.Equal(t, 4, count)
	assert.False(t, warn)
	assert.True(t, limited)
}

func TestRequestState_LimitReachedAt3rdSearch(t *testing.T) {
	t.Parallel()

	state := NewRequestState()

	// Search 1: allowed
	_, _, _, limited := state.RecordSearch("q1", "https://example.com")
	assert.False(t, limited, "first search should be allowed")

	// Search 2: allowed (max = 2)
	_, _, _, limited = state.RecordSearch("q2", "https://example.com")
	assert.False(t, limited, "second search should be allowed")

	// Search 3: blocked (exceeds MaxSearchesPerRequest=2)
	count, _, _, limited := state.RecordSearch("q3", "https://example.com")
	assert.True(t, limited, "third search should be blocked")
	assert.Equal(t, 3, count)
}

func TestRequestState_CurrentURLPreservedOnEmpty(t *testing.T) {
	t.Parallel()

	state := NewRequestState()
	state.RecordSearch("q1", "https://example.com/results")

	// Empty currentURL should not overwrite existing URL.
	state.RecordSearch("q2", "")
	assert.Equal(t, "https://example.com/results", state.CurrentURL())
}

func TestRequestState_IsLimitReached(t *testing.T) {
	t.Parallel()

	state := NewRequestState()
	assert.False(t, state.IsLimitReached(), "initially not limited")

	state.RecordSearch("q1", "")
	assert.False(t, state.IsLimitReached(), "after 1 search not limited")

	state.RecordSearch("q2", "")
	assert.False(t, state.IsLimitReached(), "after 2 searches not limited")

	state.RecordSearch("q3", "")
	assert.True(t, state.IsLimitReached(), "after 3 searches should be limited")
}

func TestErrSearchLimitReached(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, ErrSearchLimitReached, ErrSearchLimitReached)
	assert.Contains(t, ErrSearchLimitReached.Error(), "search limit")
	assert.Contains(t, ErrSearchLimitReached.Error(), "no longer available")
}

func TestRequestState_ContextRoundTrip(t *testing.T) {
	t.Parallel()

	state := NewRequestState()
	ctx := WithRequestState(context.Background(), state)

	assert.Same(t, state, RequestStateFromContext(ctx))
	assert.Nil(t, RequestStateFromContext(context.Background()))
}
