package collabview

import "time"

type CollaborationState string

const (
	CollaborationStateSolo              CollaborationState = "solo"
	CollaborationStateDelegating        CollaborationState = "delegating"
	CollaborationStateWaitingOnTeammate CollaborationState = "waiting_on_teammate"
	CollaborationStateReviewing         CollaborationState = "reviewing"
	CollaborationStateBlockedOnApproval CollaborationState = "blocked_on_approval"
	CollaborationStateRecovering        CollaborationState = "recovering"
)

type ParticipantView struct {
	Name string
}

type HandoffEdge struct {
	From      string
	To        string
	Timestamp time.Time
}

type BudgetSignal struct {
	Used      int
	Max       int
	Timestamp time.Time
}

type RecoverySignal struct {
	Action     string
	CauseClass string
	Timestamp  time.Time
}

type CollaborationView struct {
	MissionID          string
	Participants       []ParticipantView
	ActiveOwner        string
	HandoffEdges       []HandoffEdge
	CollaborationState CollaborationState
	BlockedOn          string
	BudgetSignal       *BudgetSignal
	LastRecovery       *RecoverySignal
	UpdatedAt          time.Time
}

type MissionSource struct {
	MissionID     string
	ExecutionRefs []string
	UpdatedAt     time.Time
}

type AgentRunSource struct {
	ExecutionRef     string
	RequestedAgent   string
	RuntimeCondition string
	BlockedReason    string
	UpdatedAt        time.Time
}

type RunExecutionSource struct {
	ExecutionRef      string
	CurrentStepStatus string
	UpdatedAt         time.Time
}

type DelegationSource struct {
	ExecutionRef string
	From         string
	To           string
	Timestamp    time.Time
}

type BudgetSignalSource struct {
	ExecutionRef string
	Used         int
	Max          int
	Timestamp    time.Time
}

type RecoverySignalSource struct {
	ExecutionRef string
	Action       string
	CauseClass   string
	Timestamp    time.Time
}

type ProjectionInput struct {
	Missions        []MissionSource
	AgentRuns       []AgentRunSource
	RunExecutions   []RunExecutionSource
	Delegations     []DelegationSource
	BudgetSignals   []BudgetSignalSource
	RecoverySignals []RecoverySignalSource
}
