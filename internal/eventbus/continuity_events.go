package eventbus

import "time"

// Event name constants for continuity (Phase 3) capabilities.
const (
	// EventCompactionCompleted fires after CompactionBuffer finishes a job
	// and the session's messages have been replaced with a summary entry.
	EventCompactionCompleted = "compaction.completed"

	// EventCompactionSlow fires when ContextAwareModelAdapter's sync-point
	// guard times out waiting for an in-flight compaction and proceeds with
	// the current session state.
	EventCompactionSlow = "compaction.slow"

	// EventLearningSuggestion fires when the learning engine proposes a new
	// rule/skill/preference that crosses the suggestion threshold. TUI and
	// channel adapters subscribe to render their own approval surface.
	EventLearningSuggestion = "learning.suggestion"
)

// CompactionCompletedEvent is published when a background hygiene compaction
// job replaces a prefix of a session's messages with a summary entry.
type CompactionCompletedEvent struct {
	SessionKey      string
	UpToIndex       int
	SummaryTokens   int
	ReclaimedTokens int
	Timestamp       time.Time
}

// EventName implements Event.
func (e CompactionCompletedEvent) EventName() string { return EventCompactionCompleted }

// CompactionSlowEvent is published when the turn-start sync-point guard
// exceeds its timeout waiting for an in-flight compaction. The turn proceeds
// with the current context; the event lets UIs surface the slow path to the
// user.
type CompactionSlowEvent struct {
	SessionKey string
	WaitedFor  time.Duration
	Timestamp  time.Time
}

// EventName implements Event.
func (e CompactionSlowEvent) EventName() string { return EventCompactionSlow }

// LearningSuggestionEvent is published when the learning engine proposes a
// new rule above the suggestion threshold. Subscribers render the suggestion
// as an approval prompt in their native surface (TUI chat, Slack block, etc.)
// and route the user's response through the existing approval pipeline.
type LearningSuggestionEvent struct {
	SessionKey   string
	SuggestionID string
	Pattern      string
	ProposedRule string
	Confidence   float64
	Rationale    string
	Timestamp    time.Time
}

// EventName implements Event.
func (e LearningSuggestionEvent) EventName() string { return EventLearningSuggestion }
