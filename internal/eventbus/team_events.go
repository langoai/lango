package eventbus

import "time"

// TeamFormedEvent is published when a new agent team is created.
type TeamFormedEvent struct {
	TeamID    string
	Name      string
	Goal      string
	LeaderDID string
	Members   int
}

// EventName implements Event.
func (e TeamFormedEvent) EventName() string { return "team.formed" }

// TeamDisbandedEvent is published when a team is disbanded.
type TeamDisbandedEvent struct {
	TeamID string
	Reason string
}

// EventName implements Event.
func (e TeamDisbandedEvent) EventName() string { return "team.disbanded" }

// TeamMemberJoinedEvent is published when an agent joins a team.
type TeamMemberJoinedEvent struct {
	TeamID    string
	MemberDID string
	Role      string
}

// EventName implements Event.
func (e TeamMemberJoinedEvent) EventName() string { return "team.member.joined" }

// TeamMemberLeftEvent is published when an agent leaves a team.
type TeamMemberLeftEvent struct {
	TeamID    string
	MemberDID string
	Reason    string
}

// EventName implements Event.
func (e TeamMemberLeftEvent) EventName() string { return "team.member.left" }

// TeamTaskDelegatedEvent is published when a task is sent to team workers.
type TeamTaskDelegatedEvent struct {
	TeamID   string
	ToolName string
	Workers  int
}

// EventName implements Event.
func (e TeamTaskDelegatedEvent) EventName() string { return "team.task.delegated" }

// TeamTaskCompletedEvent is published when a delegated task finishes.
type TeamTaskCompletedEvent struct {
	TeamID     string
	ToolName   string
	Successful int
	Failed     int
	Duration   time.Duration
}

// EventName implements Event.
func (e TeamTaskCompletedEvent) EventName() string { return "team.task.completed" }

// TeamConflictDetectedEvent is published when conflicting results are found.
type TeamConflictDetectedEvent struct {
	TeamID   string
	ToolName string
	Members  int
}

// EventName implements Event.
func (e TeamConflictDetectedEvent) EventName() string { return "team.conflict.detected" }

// TeamPaymentAgreedEvent is published when payment terms are negotiated.
type TeamPaymentAgreedEvent struct {
	TeamID    string
	MemberDID string
	Mode      string
	Price     string
}

// EventName implements Event.
func (e TeamPaymentAgreedEvent) EventName() string { return "team.payment.agreed" }

// TeamHealthCheckEvent is published after a team-level health sweep.
type TeamHealthCheckEvent struct {
	TeamID  string
	Healthy int
	Total   int
}

// EventName implements Event.
func (e TeamHealthCheckEvent) EventName() string { return "team.health.check" }

// TeamLeaderChangedEvent is published when a team's leader is replaced.
type TeamLeaderChangedEvent struct {
	TeamID       string
	OldLeaderDID string
	NewLeaderDID string
}

// EventName implements Event.
func (e TeamLeaderChangedEvent) EventName() string { return "team.leader.changed" }

// TeamMemberUnhealthyEvent is published when a team member misses too many health pings.
type TeamMemberUnhealthyEvent struct {
	TeamID      string
	MemberDID   string
	MemberName  string
	MissedPings int
	LastSeenAt  time.Time
}

// EventName implements Event.
func (e TeamMemberUnhealthyEvent) EventName() string { return "team.member.unhealthy" }

// TeamBudgetWarningEvent is published when a team's budget crosses a warning threshold.
type TeamBudgetWarningEvent struct {
	TeamID    string
	Threshold float64
	Spent     float64
	Budget    float64
}

// EventName implements Event.
func (e TeamBudgetWarningEvent) EventName() string { return "team.budget.warning" }

// TeamGracefulShutdownEvent is published when a team undergoes graceful shutdown.
type TeamGracefulShutdownEvent struct {
	TeamID         string
	Reason         string
	BundlesCreated int
	MembersSettled int
}

// EventName implements Event.
func (e TeamGracefulShutdownEvent) EventName() string { return "team.graceful.shutdown" }
