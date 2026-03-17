package cron

import "time"

// Job represents a scheduled cron job in the domain layer.
type Job struct {
	ID           string
	Name         string
	ScheduleType string // "at" | "every" | "cron"
	Schedule     string
	Prompt       string
	SessionMode  string // "isolated" | "main"
	DeliverTo    []string
	Timezone     string
	Enabled      bool
	Timeout      time.Duration // per-job timeout (0 = use default)
	LastRunAt    *time.Time
	NextRunAt    *time.Time
	CreatedAt    time.Time
}

// JobResult holds the outcome of a single job execution.
type JobResult struct {
	JobID     string
	JobName   string
	Response  string
	Error     error
	StartedAt time.Time
	Duration  time.Duration
}

// HistoryEntry represents a persisted execution record for a cron job.
type HistoryEntry struct {
	ID           string
	JobID        string
	JobName      string
	Status       string // "running" | "completed" | "failed"
	Prompt       string
	Result       string
	ErrorMessage string
	TokensUsed   int
	StartedAt    time.Time
	CompletedAt  *time.Time
}
