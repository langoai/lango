package background

import (
	"context"
	"sync"
	"time"
)

// Status represents the lifecycle state of a background task.
type Status int

const (
	Pending Status = iota + 1
	Running
	Done
	Failed
	Cancelled
)

// Valid reports whether s is a known task status.
func (s Status) Valid() bool {
	switch s {
	case Pending, Running, Done, Failed, Cancelled:
		return true
	}
	return false
}

// Values returns all known task statuses.
func (s Status) Values() []Status {
	return []Status{Pending, Running, Done, Failed, Cancelled}
}

// String returns the human-readable name of the status.
func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Running:
		return "running"
	case Done:
		return "done"
	case Failed:
		return "failed"
	case Cancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Task represents a background execution unit.
type Task struct {
	ID            string
	Status        Status
	Prompt        string
	Result        string
	Error         string
	RetryKey      string // canonical retry identity for resubmitted work
	AttemptCount  int
	NextRetryAt   time.Time
	OriginChannel string // channel that initiated the request (e.g. "telegram", "slack")
	OriginSession string // original session key
	StartedAt     time.Time
	CompletedAt   time.Time
	TokensUsed    int
	mu            sync.RWMutex
	cancelFn      context.CancelFunc
}

// TaskSnapshot is an immutable copy of a Task, safe for concurrent reading.
type TaskSnapshot struct {
	ID            string    `json:"id"`
	Status        Status    `json:"status"`
	StatusText    string    `json:"status_text"`
	Prompt        string    `json:"prompt"`
	Result        string    `json:"result"`
	Error         string    `json:"error,omitempty"`
	RetryKey      string    `json:"retry_key,omitempty"`
	AttemptCount  int       `json:"attempt_count"`
	NextRetryAt   time.Time `json:"next_retry_at,omitempty"`
	OriginChannel string    `json:"origin_channel"`
	OriginSession string    `json:"origin_session"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	TokensUsed    int       `json:"tokens_used"`
}

// SetRunning transitions the task to the Running state and records the start time.
func (t *Task) SetRunning() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = Running
	t.AttemptCount++
	t.StartedAt = time.Now()
	t.CompletedAt = time.Time{}
	t.NextRetryAt = time.Time{}
	t.Result = ""
	t.Error = ""
}

// Complete transitions the task to the Done state with the given result.
// If the task is already Cancelled, the transition is skipped to preserve
// the cancellation status.
func (t *Task) Complete(result string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Cancelled {
		return
	}
	t.Status = Done
	t.Result = result
	t.Error = ""
	t.NextRetryAt = time.Time{}
	t.CompletedAt = time.Now()
}

// Fail transitions the task to the Failed state with the given error message.
// If the task is already Cancelled, the transition is skipped to preserve
// the cancellation status.
func (t *Task) Fail(errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Cancelled {
		return
	}
	t.Status = Failed
	t.Error = errMsg
	t.Result = ""
	t.CompletedAt = time.Now()
}

// Cancel transitions the task to the Cancelled state and invokes the cancel function.
func (t *Task) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = Cancelled
	t.NextRetryAt = time.Time{}
	t.CompletedAt = time.Now()
	if t.cancelFn != nil {
		t.cancelFn()
	}
}

// ScheduleRetry records when the next retry should occur.
func (t *Task) ScheduleRetry(nextRetryAt time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Cancelled {
		return
	}
	t.NextRetryAt = nextRetryAt
}

// SetRetryKey records the canonical retry identity for this task.
func (t *Task) SetRetryKey(retryKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.RetryKey = retryKey
}

// Snapshot returns an immutable copy of the task's current state.
func (t *Task) Snapshot() TaskSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return TaskSnapshot{
		ID:            t.ID,
		Status:        t.Status,
		StatusText:    t.Status.String(),
		Prompt:        t.Prompt,
		Result:        t.Result,
		Error:         t.Error,
		RetryKey:      t.RetryKey,
		AttemptCount:  t.AttemptCount,
		NextRetryAt:   t.NextRetryAt,
		OriginChannel: t.OriginChannel,
		OriginSession: t.OriginSession,
		StartedAt:     t.StartedAt,
		CompletedAt:   t.CompletedAt,
		TokensUsed:    t.TokensUsed,
	}
}
