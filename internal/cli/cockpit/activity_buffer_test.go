package cockpit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
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
