package cockpit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
)

const missionActivityCapacity = 200

type MissionActivityKind string

const (
	MissionActivityGeneric    MissionActivityKind = "generic"
	MissionActivityContinuity MissionActivityKind = "continuity"
	MissionActivityLearning   MissionActivityKind = "learning"
	MissionActivityRuntime    MissionActivityKind = "runtime"
	MissionActivityUser       MissionActivityKind = "user"
	MissionActivityTurn       MissionActivityKind = "turn"
)

// MissionActivityItem is one deterministic activity row for the cockpit
// timeline projection.
type MissionActivityItem struct {
	Kind       MissionActivityKind
	SessionKey string
	Summary    string
	Timestamp  time.Time
}

// MissionActivityBuffer keeps recent cockpit-lifetime activity items.
type MissionActivityBuffer struct {
	mu    sync.RWMutex
	items []MissionActivityItem
}

func NewMissionActivityBuffer() *MissionActivityBuffer {
	return &MissionActivityBuffer{}
}

func (b *MissionActivityBuffer) Append(item MissionActivityItem) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.items) >= missionActivityCapacity {
		copy(b.items, b.items[1:])
		b.items[len(b.items)-1] = item
		return
	}
	b.items = append(b.items, item)
}

func (b *MissionActivityBuffer) Snapshot() []MissionActivityItem {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.items) == 0 {
		return nil
	}
	out := make([]MissionActivityItem, len(b.items))
	copy(out, b.items)
	return out
}

func (b *MissionActivityBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = nil
}

func newCompactionCompletedActivity(event eventbus.CompactionCompletedEvent) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityContinuity,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Context compacted, reclaimed %d tokens", event.ReclaimedTokens),
		Timestamp:  event.Timestamp,
	}
}

func newCompactionSlowActivity(event eventbus.CompactionSlowEvent) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityContinuity,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Compaction slow path after %s", event.WaitedFor.Round(time.Second)),
		Timestamp:  event.Timestamp,
	}
}

func newLearningSuggestionActivity(event eventbus.LearningSuggestionEvent) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityLearning,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Learning suggestion %.0f%%: %s", event.Confidence*100, event.ProposedRule),
		Timestamp:  event.Timestamp,
	}
}

func newModeChangedActivity(event eventbus.ModeChangedEvent, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityGeneric,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Mode changed from %q to %q", emptyMode(event.OldMode), emptyMode(event.NewMode)),
		Timestamp:  now,
	}
}

func newTurnCompletedActivity(event eventbus.TurnCompletedEvent, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityGeneric,
		SessionKey: event.SessionKey,
		Summary:    "Turn completed",
		Timestamp:  now,
	}
}

func newPolicyDecisionActivity(event eventbus.PolicyDecisionEvent, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityGeneric,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Policy %s for %s", event.Verdict, event.Command),
		Timestamp:  now,
	}
}

func newAlertActivity(event eventbus.AlertEvent) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityGeneric,
		SessionKey: event.SessionKey,
		Summary:    fmt.Sprintf("Alert %s: %s", event.Severity, event.Message),
		Timestamp:  event.Timestamp,
	}
}

func newRunLedgerMirrorFailureActivity(event eventbus.RunLedgerMirrorFailureEvent, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:      MissionActivityGeneric,
		Summary:   fmt.Sprintf("RunLedger mirror failure for %s during %s", event.Target, event.Phase),
		Timestamp: now,
	}
}

func newChannelMessageActivity(msg chat.ChannelMessageMsg) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityRuntime,
		SessionKey: msg.SessionKey,
		Summary:    fmt.Sprintf("%s message from %s: %s", msg.Channel, msg.SenderName, strings.TrimSpace(msg.Text)),
		Timestamp:  msg.Timestamp,
	}
}

func newDelegationActivity(msg chat.DelegationMsg, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:      MissionActivityRuntime,
		Summary:   fmt.Sprintf("Delegated from %s to %s (%s)", msg.From, msg.To, msg.Reason),
		Timestamp: now,
	}
}

func newBudgetWarningActivity(msg chat.BudgetWarningMsg, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:      MissionActivityRuntime,
		Summary:   fmt.Sprintf("Delegation budget warning %d/%d", msg.Used, msg.Max),
		Timestamp: now,
	}
}

func newRecoveryActivity(msg chat.RecoveryMsg, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:      MissionActivityRuntime,
		Summary:   fmt.Sprintf("Recovery %s after %s (attempt %d)", msg.Action, msg.CauseClass, msg.Attempt),
		Timestamp: now,
	}
}

func newUserSubmissionActivity(sessionKey, input string, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityUser,
		SessionKey: sessionKey,
		Summary:    fmt.Sprintf("User submitted: %s", strings.TrimSpace(input)),
		Timestamp:  now,
	}
}

func newTurnSummaryActivity(sessionKey string, msg chat.TurnTokenUsageMsg, now time.Time) MissionActivityItem {
	return MissionActivityItem{
		Kind:       MissionActivityTurn,
		SessionKey: sessionKey,
		Summary:    fmt.Sprintf("Turn summary: %d total tokens (%d in / %d out)", msg.TotalTokens, msg.InputTokens, msg.OutputTokens),
		Timestamp:  now,
	}
}

func emptyMode(name string) string {
	if strings.TrimSpace(name) == "" {
		return "none"
	}
	return name
}
