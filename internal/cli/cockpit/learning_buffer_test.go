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

func TestLearningSuggestionBufferFind(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	buf := NewLearningSuggestionBuffer(func() time.Time { return now })

	buf.Append(eventbus.LearningSuggestionEvent{
		SessionKey:   "sess-1",
		SuggestionID: "s-1",
		Pattern:      "retry\x1b[31m timeout\n",
		ProposedRule: "Use\x1b[31m bounded\nretry",
		Rationale:    "Pattern\x1b[31m repeated\n",
		Timestamp:    now,
	})

	found := buf.Find("s-1")
	require.NotNil(t, found)
	assert.Equal(t, "retry timeout", found.Pattern)
	assert.Equal(t, "Use bounded retry", found.ProposedRule)
	assert.Equal(t, "Pattern repeated", found.Rationale)
	assert.NotContains(t, found.Pattern, "\x1b")

	assert.Nil(t, buf.Find("missing"))
}
