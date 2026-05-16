package cockpit

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/turnrunner"
)

func TestMissionActivityBufferRetention(t *testing.T) {
	buf := NewMissionActivityBuffer()

	for i := 0; i < 205; i++ {
		buf.Append(MissionActivityItem{
			Kind:       MissionActivityGeneric,
			SessionKey: "sess-1",
			Summary:    time.Unix(int64(i), 0).UTC().Format(time.RFC3339),
			Timestamp:  time.Unix(int64(i), 0).UTC(),
		})
	}

	items := buf.Snapshot()
	require.Len(t, items, 200)
	assert.Equal(t, time.Unix(5, 0).UTC().Format(time.RFC3339), items[0].Summary)
	assert.Equal(t, time.Unix(204, 0).UTC().Format(time.RFC3339), items[len(items)-1].Summary)
}

func TestMissionActivityBufferAppendCompactsSummary(t *testing.T) {
	buf := NewMissionActivityBuffer()
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	buf.Append(MissionActivityItem{
		Kind:      MissionActivityAssistant,
		Summary:   "  First\x1b[31m line\n\nSecond   line\twith  extra spacing  ",
		Timestamp: now,
	})

	longText := "Assistant reply: " + strings.Repeat("very long summary ", 20)
	buf.Append(MissionActivityItem{
		Kind:      MissionActivityAssistant,
		Summary:   longText,
		Timestamp: now.Add(time.Second),
	})

	items := buf.Snapshot()
	require.Len(t, items, 2)
	assert.Equal(t, "First line Second line with extra spacing", items[0].Summary)
	assert.NotContains(t, items[0].Summary, "\x1b")
	assert.Len(t, []rune(items[1].Summary), missionActivitySummaryMaxRunes)
	assert.True(t, strings.HasSuffix(items[1].Summary, "..."))
	assert.NotContains(t, items[1].Summary, "\n")
}

func TestNewAssistantSummaryActivity(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	t.Run("success summary", func(t *testing.T) {
		item, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{
			Result: turnrunner.Result{
				Outcome: "success",
				Summary: "Answer completed",
			},
		}, now)
		require.True(t, ok)
		assert.Equal(t, MissionActivityAssistant, item.Kind)
		assert.Equal(t, "sess-1", item.SessionKey)
		assert.Equal(t, "Assistant reply: Answer completed", item.Summary)
		assert.Equal(t, now, item.Timestamp)
	})

	t.Run("failure summary", func(t *testing.T) {
		item, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{
			Result: turnrunner.Result{
				Outcome:     "timeout",
				UserMessage: "Request timed out",
				Summary:     "request\x1b[31m exceeded\nlimit",
			},
		}, now)
		require.True(t, ok)
		assert.Equal(t, "Turn timeout: request exceeded limit", item.Summary)
		assert.NotContains(t, item.Summary, "\x1b")
	})

	t.Run("prefers summary and sanitizes before append", func(t *testing.T) {
		item, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{
			Result: turnrunner.Result{
				Outcome:      "success",
				ResponseText: "Response body fallback",
				Summary:      "Line \x1b[31mone\n\nLine two",
			},
		}, now)
		require.True(t, ok)
		assert.Equal(t, "Assistant reply: Line one Line two", item.Summary)
		assert.NotContains(t, item.Summary, "\x1b")

		buf := NewMissionActivityBuffer()
		buf.Append(item)
		items := buf.Snapshot()
		require.Len(t, items, 1)
		assert.Equal(t, "Assistant reply: Line one Line two", items[0].Summary)
		assert.NotContains(t, items[0].Summary, "\x1b")
	})

	t.Run("empty summary skipped", func(t *testing.T) {
		_, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{}, now)
		assert.False(t, ok)
	})
}

func TestMissionActivityContinuityEventAppend(t *testing.T) {
	bus := eventbus.New()
	learning := NewLearningSuggestionBuffer(nil)
	activity := NewMissionActivityBuffer()

	SubscribeMissionControlEvents(bus, "sess-1", learning, activity)

	bus.Publish(eventbus.CompactionCompletedEvent{
		SessionKey:      "sess-1",
		ReclaimedTokens: 42,
		Timestamp:       time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	})

	items := activity.Snapshot()
	require.Len(t, items, 1)
	assert.Equal(t, MissionActivityContinuity, items[0].Kind)
	assert.Contains(t, items[0].Summary, "reclaimed 42 tokens")
}

func TestMissionActivityShellRuntimeEventAppend(t *testing.T) {
	mock := &mockChild{}
	buf := NewMissionActivityBuffer()
	m := newTestModel(mock)
	m.activityBuffer = buf

	m.handleChannelMessage(chat.ChannelMessageMsg{
		Channel:    "telegram",
		SessionKey: "telegram:1:2",
		SenderName: "alice",
		Text:       "hello",
		Timestamp:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	})
	m.handleDelegation(chat.DelegationMsg{From: "operator", To: "researcher", Reason: "search"})
	m.handleBudgetWarning(chat.BudgetWarningMsg{Used: 9, Max: 10})
	m.handleRecovery(chat.RecoveryMsg{CauseClass: "rate_limit", Action: "retry", Attempt: 2})

	items := buf.Snapshot()
	require.Len(t, items, 4)
	assert.Contains(t, items[0].Summary, "alice")
	assert.Contains(t, items[1].Summary, "researcher")
	assert.Contains(t, items[2].Summary, "9/10")
	assert.Contains(t, items[3].Summary, "retry")
}

func TestMissionActivityIgnoresRunLedgerMirrorFailureWithoutSessionScope(t *testing.T) {
	bus := eventbus.New()
	learning := NewLearningSuggestionBuffer(nil)
	activity := NewMissionActivityBuffer()

	SubscribeMissionControlEvents(bus, "sess-1", learning, activity)

	bus.Publish(eventbus.RunLedgerMirrorFailureEvent{
		Target: "agent_run_projection",
		Phase:  "append_journal",
		RunID:  "run-1",
		Error:  "disk full",
	})

	assert.Empty(t, activity.Snapshot())
}
