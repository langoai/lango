package cockpit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
)

// mockSender is defined in channelbridge_test.go (same package).

func TestRuntimeTracker_TokenAccumulation(t *testing.T) {
	tests := []struct {
		give           string
		giveEvents     []eventbus.TokenUsageEvent
		wantInput      int64
		wantOutput     int64
		wantTotal      int64
		wantCache      int64
	}{
		{
			give: "single event",
			giveEvents: []eventbus.TokenUsageEvent{
				{SessionKey: "sess-1", InputTokens: 10, OutputTokens: 20, TotalTokens: 30, CacheTokens: 5},
			},
			wantInput:  10,
			wantOutput: 20,
			wantTotal:  30,
			wantCache:  5,
		},
		{
			give: "three events summed",
			giveEvents: []eventbus.TokenUsageEvent{
				{SessionKey: "sess-1", InputTokens: 10, OutputTokens: 20, TotalTokens: 30, CacheTokens: 5},
				{SessionKey: "sess-1", InputTokens: 100, OutputTokens: 200, TotalTokens: 300, CacheTokens: 50},
				{SessionKey: "sess-1", InputTokens: 1, OutputTokens: 2, TotalTokens: 3, CacheTokens: 0},
			},
			wantInput:  111,
			wantOutput: 222,
			wantTotal:  333,
			wantCache:  55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			bus := eventbus.New()
			sender := &mockSender{}
			tracker := NewRuntimeTracker(bus, sender, "sess-1")
			tracker.StartTurn() // must be active to accumulate

			for _, e := range tt.giveEvents {
				bus.Publish(e)
			}

			snap := tracker.FlushTurnTokens()
			assert.Equal(t, tt.wantInput, snap.InputTokens)
			assert.Equal(t, tt.wantOutput, snap.OutputTokens)
			assert.Equal(t, tt.wantTotal, snap.TotalTokens)
			assert.Equal(t, tt.wantCache, snap.CacheTokens)
		})
	}
}

func TestRuntimeTracker_TokenIgnoredWhenTurnInactive(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	// Do NOT call StartTurn — tokens should be ignored.

	bus.Publish(eventbus.TokenUsageEvent{
		SessionKey:  "",
		InputTokens: 100, OutputTokens: 200, TotalTokens: 300,
	})

	snap := tracker.FlushTurnTokens()
	assert.Equal(t, int64(0), snap.TotalTokens, "tokens should be ignored when turn is inactive")
}

func TestRuntimeTracker_TokenSessionFilter(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	tracker.StartTurn()

	// Publish event with different session key — should be filtered out.
	bus.Publish(eventbus.TokenUsageEvent{
		SessionKey:   "sess-other",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		CacheTokens:  50,
	})

	snap := tracker.FlushTurnTokens()
	assert.Equal(t, int64(0), snap.InputTokens)
	assert.Equal(t, int64(0), snap.OutputTokens)
	assert.Equal(t, int64(0), snap.TotalTokens)
	assert.Equal(t, int64(0), snap.CacheTokens)
}

func TestRuntimeTracker_TokenEmptySessionKeyAccepted(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	tracker.StartTurn()

	// The production publisher (wireModelAdapterTokenUsage) does not set
	// SessionKey, so events arrive with SessionKey="". These must be
	// accepted as belonging to the local session.
	bus.Publish(eventbus.TokenUsageEvent{
		SessionKey:   "",
		InputTokens:  42,
		OutputTokens: 84,
		TotalTokens:  126,
		CacheTokens:  10,
	})

	snap := tracker.FlushTurnTokens()
	assert.Equal(t, int64(42), snap.InputTokens)
	assert.Equal(t, int64(84), snap.OutputTokens)
	assert.Equal(t, int64(126), snap.TotalTokens)
	assert.Equal(t, int64(10), snap.CacheTokens)
}

func TestRuntimeTracker_TokenFlushReset(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	tracker.StartTurn()

	bus.Publish(eventbus.TokenUsageEvent{
		SessionKey:   "sess-1",
		InputTokens:  10,
		OutputTokens: 20,
		TotalTokens:  30,
		CacheTokens:  5,
	})

	// First flush returns accumulated values.
	first := tracker.FlushTurnTokens()
	assert.Equal(t, int64(30), first.TotalTokens)

	// Second flush returns zero.
	second := tracker.FlushTurnTokens()
	assert.Equal(t, int64(0), second.InputTokens)
	assert.Equal(t, int64(0), second.OutputTokens)
	assert.Equal(t, int64(0), second.TotalTokens)
	assert.Equal(t, int64(0), second.CacheTokens)
}

func TestRuntimeTracker_RecoveryForwarding(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	_ = tracker // keep reference alive

	bus.Publish(agentrt.RecoveryDecisionEvent{
		CauseClass: "timeout",
		Action:     "retry",
		Attempt:    2,
		Backoff:    3 * time.Second,
		SessionKey: "sess-1",
	})

	require.Len(t, sender.msgs, 1)
	msg, ok := sender.msgs[0].(chat.RecoveryMsg)
	require.True(t, ok, "expected chat.RecoveryMsg, got %T", sender.msgs[0])

	assert.Equal(t, "timeout", msg.CauseClass)
	assert.Equal(t, "retry", msg.Action)
	assert.Equal(t, 2, msg.Attempt)
	assert.Equal(t, 3*time.Second, msg.Backoff)
}

func TestRuntimeTracker_RecoverySessionFilter(t *testing.T) {
	bus := eventbus.New()
	sender := &mockSender{}
	tracker := NewRuntimeTracker(bus, sender, "sess-1")
	_ = tracker

	bus.Publish(agentrt.RecoveryDecisionEvent{
		CauseClass: "timeout",
		Action:     "retry",
		Attempt:    1,
		Backoff:    time.Second,
		SessionKey: "sess-other",
	})

	assert.Empty(t, sender.msgs)
}

func TestRuntimeTracker_RecordDelegation(t *testing.T) {
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	tracker.RecordDelegation("agent-b")

	snap := tracker.Snapshot()
	assert.Equal(t, 1, snap.DelegationCount)
	assert.Equal(t, "agent-b", snap.ActiveAgent)
	assert.True(t, snap.IsRunning)

	// Second delegation increments count and updates active agent.
	tracker.RecordDelegation("agent-c")

	snap = tracker.Snapshot()
	assert.Equal(t, 2, snap.DelegationCount)
	assert.Equal(t, "agent-c", snap.ActiveAgent)
}

func TestRuntimeTracker_SetActiveAgent(t *testing.T) {
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	tracker.RecordDelegation("operator")
	tracker.SetActiveAgent("lango-orchestrator")

	snap := tracker.Snapshot()
	assert.Equal(t, "lango-orchestrator", snap.ActiveAgent)
	assert.Equal(t, 1, snap.DelegationCount, "SetActiveAgent should not increment counter")
}

func TestRuntimeTracker_IsRunningFromTurnActive(t *testing.T) {
	tracker := NewRuntimeTracker(nil, nil, "sess-1")

	// Before StartTurn — not running.
	assert.False(t, tracker.Snapshot().IsRunning)

	// After StartTurn — running even without delegation.
	tracker.StartTurn()
	assert.True(t, tracker.Snapshot().IsRunning)

	// After ResetTurn — not running.
	tracker.ResetTurn()
	assert.False(t, tracker.Snapshot().IsRunning)
}

func TestRuntimeTracker_ResetTurn(t *testing.T) {
	tracker := NewRuntimeTracker(nil, nil, "sess-1")
	tracker.StartTurn()
	tracker.RecordDelegation("agent-b")

	// Verify pre-reset state.
	snap := tracker.Snapshot()
	assert.Equal(t, 1, snap.DelegationCount)
	assert.Equal(t, "agent-b", snap.ActiveAgent)
	assert.True(t, snap.IsRunning)

	tracker.ResetTurn()

	snap = tracker.Snapshot()
	assert.Equal(t, 0, snap.DelegationCount)
	assert.Equal(t, "", snap.ActiveAgent)
	assert.False(t, snap.IsRunning)
}

func TestRuntimeTracker_NilBus(t *testing.T) {
	// Must not panic with nil bus.
	tracker := NewRuntimeTracker(nil, &mockSender{}, "sess-1")
	assert.NotNil(t, tracker)

	// Methods work without bus.
	snap := tracker.FlushTurnTokens()
	assert.Equal(t, int64(0), snap.TotalTokens)

	tracker.RecordDelegation("agent-x")
	status := tracker.Snapshot()
	assert.Equal(t, "agent-x", status.ActiveAgent)
}
