package cockpit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
)

func TestLearningSuggestionBufferTTL(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	buf := NewLearningSuggestionBuffer(func() time.Time { return now })

	buf.Append(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "keep",
		ProposedRule: "Keep this",
		Timestamp:    now.Add(-10 * time.Minute),
	})
	buf.Append(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "expire",
		ProposedRule: "Expire this",
		Timestamp:    now.Add(-31 * time.Minute),
	})

	items := buf.Snapshot()
	require.Len(t, items, 1)
	assert.Equal(t, "keep", items[0].SuggestionID)
}

func TestLearningSuggestionBufferDismiss(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	buf := NewLearningSuggestionBuffer(func() time.Time { return now })

	buf.Append(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "s-1",
		ProposedRule: "Rule one",
		Timestamp:    now,
	})
	buf.Append(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "s-2",
		ProposedRule: "Rule two",
		Timestamp:    now,
	})

	buf.Dismiss("s-1")

	items := buf.Snapshot()
	require.Len(t, items, 1)
	assert.Equal(t, "s-2", items[0].SuggestionID)
}
