package eventbus

import "time"

// Event name constants for team domain events.
const (
	EventTeamFormed          = "team.formed"
	EventTeamDisbanded       = "team.disbanded"
	EventTeamMemberJoined    = "team.member.joined"
	EventTeamMemberLeft      = "team.member.left"
	EventTeamTaskDelegated   = "team.task.delegated"
	EventTeamTaskCompleted   = "team.task.completed"
	EventTeamConflictDetected = "team.conflict.detected"
	EventTeamPaymentAgreed   = "team.payment.agreed"
	EventTeamHealthCheck     = "team.health.check"
	EventTeamLeaderChanged   = "team.leader.changed"
	EventTeamMemberUnhealthy = "team.member.unhealthy"
	EventTeamBudgetWarning   = "team.budget.warning"
	EventTeamGracefulShutdown = "team.graceful.shutdown"
)

// TeamFormedEvent is published when a new agent team is created.
type TeamFormedEvent struct {
	TeamID    string
	Name      string
	Goal      string
	LeaderDID string
	Members   int
}

// EventName implements Event.
func (e TeamFormedEvent) EventName() string { return EventTeamFormed }

// TeamDisbandedEvent is published when a team is disbanded.
type TeamDisbandedEvent struct {
	TeamID string
	Reason string
}

// EventName implements Event.
func (e TeamDisbandedEvent) EventName() string { return EventTeamDisbanded }

// TeamMemberJoinedEvent is published when an agent joins a team.
type TeamMemberJoinedEvent struct {
	TeamID    string
	MemberDID string
	Role      string
}

// EventName implements Event.
func (e TeamMemberJoinedEvent) EventName() string { return EventTeamMemberJoined }

// TeamMemberLeftEvent is published when an agent leaves a team.
type TeamMemberLeftEvent struct {
	TeamID    string
	MemberDID string
	Reason    string
}

// EventName implements Event.
func (e TeamMemberLeftEvent) EventName() string { return EventTeamMemberLeft }

// TeamTaskDelegatedEvent is published when a task is sent to team workers.
type TeamTaskDelegatedEvent struct {
	TeamID   string
	ToolName string
	Workers  int
}

// EventName implements Event.
func (e TeamTaskDelegatedEvent) EventName() string { return EventTeamTaskDelegated }

// TeamTaskCompletedEvent is published when a delegated task finishes.
type TeamTaskCompletedEvent struct {
	TeamID     string
	ToolName   string
	Successful int
	Failed     int
	Duration   time.Duration
}

// EventName implements Event.
func (e TeamTaskCompletedEvent) EventName() string { return EventTeamTaskCompleted }

// TeamConflictDetectedEvent is published when conflicting results are found.
type TeamConflictDetectedEvent struct {
	TeamID   string
	ToolName string
	Members  int
}

// EventName implements Event.
func (e TeamConflictDetectedEvent) EventName() string { return EventTeamConflictDetected }

// TeamPaymentAgreedEvent is published when payment terms are negotiated.
type TeamPaymentAgreedEvent struct {
	TeamID    string
	MemberDID string
	Mode      string
	Price     string
}

// EventName implements Event.
func (e TeamPaymentAgreedEvent) EventName() string { return EventTeamPaymentAgreed }

// TeamHealthCheckEvent is published after a team-level health sweep.
type TeamHealthCheckEvent struct {
	TeamID  string
	Healthy int
	Total   int
}

// EventName implements Event.
func (e TeamHealthCheckEvent) EventName() string { return EventTeamHealthCheck }

// TeamLeaderChangedEvent is published when a team's leader is replaced.
type TeamLeaderChangedEvent struct {
	TeamID       string
	OldLeaderDID string
	NewLeaderDID string
}

// EventName implements Event.
func (e TeamLeaderChangedEvent) EventName() string { return EventTeamLeaderChanged }

// TeamMemberUnhealthyEvent is published when a team member misses too many health pings.
type TeamMemberUnhealthyEvent struct {
	TeamID      string
	MemberDID   string
	MemberName  string
	MissedPings int
	LastSeenAt  time.Time
}

// EventName implements Event.
func (e TeamMemberUnhealthyEvent) EventName() string { return EventTeamMemberUnhealthy }

// TeamBudgetWarningEvent is published when a team's budget crosses a warning threshold.
type TeamBudgetWarningEvent struct {
	TeamID    string
	Threshold float64
	Spent     float64
	Budget    float64
}

// EventName implements Event.
func (e TeamBudgetWarningEvent) EventName() string { return EventTeamBudgetWarning }

// TeamGracefulShutdownEvent is published when a team undergoes graceful shutdown.
type TeamGracefulShutdownEvent struct {
	TeamID         string
	Reason         string
	BundlesCreated int
	MembersSettled int
}

// EventName implements Event.
func (e TeamGracefulShutdownEvent) EventName() string { return EventTeamGracefulShutdown }
