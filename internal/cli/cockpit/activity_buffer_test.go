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

func TestMissionActivityBufferResetAndEmptySnapshot(t *testing.T) {
	buf := NewMissionActivityBuffer()
	assert.Nil(t, buf.Snapshot())

	buf.Append(MissionActivityItem{Summary: "visible"})
	require.Len(t, buf.Snapshot(), 1)

	buf.Reset()

	assert.Nil(t, buf.Snapshot())
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

	t.Run("falls back to response text", func(t *testing.T) {
		item, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{
			Result: turnrunner.Result{
				Outcome:      "success",
				ResponseText: "Response body fallback",
			},
		}, now)
		require.True(t, ok)
		assert.Equal(t, "Assistant reply: Response body fallback", item.Summary)
	})

	t.Run("empty summary skipped", func(t *testing.T) {
		_, ok := newAssistantSummaryActivity("sess-1", chat.DoneMsg{}, now)
		assert.False(t, ok)
	})
}

func TestMissionActivityConstructors(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		item        MissionActivityItem
		wantKind    MissionActivityKind
		wantSession string
		wantSummary string
	}{
		{
			name: "compaction slow",
			item: newCompactionSlowActivity(eventbus.CompactionSlowEvent{
				SessionKey: "sess-1",
				WaitedFor:  1500 * time.Millisecond,
				Timestamp:  now,
			}),
			wantKind:    MissionActivityContinuity,
			wantSession: "sess-1",
			wantSummary: "Compaction slow path after 2s",
		},
		{
			name: "learning suggestion",
			item: newLearningSuggestionActivity(eventbus.LearningSuggestionEvent{
				SessionKey:   "sess-1",
				Confidence:   0.87,
				ProposedRule: "Prefer smaller slices",
				Timestamp:    now,
			}),
			wantKind:    MissionActivityLearning,
			wantSession: "sess-1",
			wantSummary: "Learning suggestion 87%: Prefer smaller slices",
		},
		{
			name: "mode changed empty old mode",
			item: newModeChangedActivity(eventbus.ModeChangedEvent{
				SessionKey: "sess-1",
				OldMode:    " ",
				NewMode:    "agent",
			}, now),
			wantKind:    MissionActivityGeneric,
			wantSession: "sess-1",
			wantSummary: "Mode changed from \"none\" to \"agent\"",
		},
		{
			name:        "turn completed",
			item:        newTurnCompletedActivity(eventbus.TurnCompletedEvent{SessionKey: "sess-1"}, now),
			wantKind:    MissionActivityGeneric,
			wantSession: "sess-1",
			wantSummary: "Turn completed",
		},
		{
			name: "policy decision",
			item: newPolicyDecisionActivity(eventbus.PolicyDecisionEvent{
				SessionKey: "sess-1",
				Verdict:    "allow",
				Command:    "go test ./...",
			}, now),
			wantKind:    MissionActivityGeneric,
			wantSession: "sess-1",
			wantSummary: "Policy allow for go test ./...",
		},
		{
			name: "alert",
			item: newAlertActivity(eventbus.AlertEvent{
				SessionKey: "sess-1",
				Severity:   "warning",
				Message:    "budget near limit",
				Timestamp:  now,
			}),
			wantKind:    MissionActivityGeneric,
			wantSession: "sess-1",
			wantSummary: "Alert warning: budget near limit",
		},
		{
			name:        "run ledger mirror failure",
			item:        newRunLedgerMirrorFailureActivity(eventbus.RunLedgerMirrorFailureEvent{Target: "projection", Phase: "sync"}, now),
			wantKind:    MissionActivityGeneric,
			wantSummary: "RunLedger mirror failure for projection during sync",
		},
		{
			name:        "user submission trims input",
			item:        newUserSubmissionActivity("sess-1", "  hello\n", now),
			wantKind:    MissionActivityUser,
			wantSession: "sess-1",
			wantSummary: "User submitted: hello",
		},
		{
			name: "turn summary",
			item: newTurnSummaryActivity("sess-1", chat.TurnTokenUsageMsg{
				TotalTokens:  15,
				InputTokens:  10,
				OutputTokens: 5,
			}, now),
			wantKind:    MissionActivityTurn,
			wantSession: "sess-1",
			wantSummary: "Turn summary: 15 total tokens (10 in / 5 out)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKind, tt.item.Kind)
			assert.Equal(t, tt.wantSession, tt.item.SessionKey)
			assert.Equal(t, tt.wantSummary, tt.item.Summary)
			assert.Equal(t, now, tt.item.Timestamp)
		})
	}
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
