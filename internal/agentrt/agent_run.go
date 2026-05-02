package agentrt

import (
	"context"
	"time"
)

// AgentRunStatus represents the lifecycle status of an agent run.
type AgentRunStatus string

const (
	AgentRunSpawned   AgentRunStatus = "spawned"
	AgentRunRunning   AgentRunStatus = "running"
	AgentRunCompleted AgentRunStatus = "completed"
	AgentRunFailed    AgentRunStatus = "failed"
	AgentRunCancelled AgentRunStatus = "cancelled"
)

// AgentRunCondition represents a projected runtime condition for an agent run.
type AgentRunCondition string

const (
	AgentRunConditionNone                   AgentRunCondition = ""
	AgentRunConditionBlockedWaitingApproval AgentRunCondition = "blocked_waiting_approval"
	AgentRunConditionBlockedWaitingMessage  AgentRunCondition = "blocked_waiting_message"
	AgentRunConditionWaitingOnTeammate      AgentRunCondition = "waiting_on_teammate"
	AgentRunConditionResuming               AgentRunCondition = "resuming"
	AgentRunConditionOrphaned               AgentRunCondition = "orphaned"
	AgentRunConditionRecovering             AgentRunCondition = "recovering"
)

// isTerminal reports whether the status is a final state that cannot be overwritten.
func (s AgentRunStatus) isTerminal() bool {
	switch s {
	case AgentRunCompleted, AgentRunFailed, AgentRunCancelled:
		return true
	}
	return false
}

// AgentRun tracks a spawned agent's lifecycle.
// ID is unified with the background manager's task ID via D1a projection.
type AgentRun struct {
	ID               string // unified with bgManager task ID
	ParentID         string // parent agent/session
	RequestedAgent   string // advisory target specialist (not guaranteed routing)
	Instruction      string
	Status           AgentRunStatus
	ChildSession     string // child session key
	Result           string
	Error            string
	SpawnDepth       int
	AllowedTools     []string // tool restrictions for this run
	SpawnReason      string
	RuntimeCondition AgentRunCondition
	BlockedReason    string
	GrantRequestID   string
	GrantAttempt     int
	GrantState       string
	WaitingOnRunID   string
	RecoveryState    string
	CreatedAt        time.Time
	CompletedAt      time.Time
	CancelFn         context.CancelFunc `json:"-"`
}
