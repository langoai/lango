package cockpit

import "time"

// MissionKind identifies whether a projected mission is active or proposed.
type MissionKind string

const (
	MissionKindUnknown  MissionKind = "unknown"
	MissionKindActive   MissionKind = "active"
	MissionKindProposed MissionKind = "proposed"
)

// MissionStatus is the limited mission status space for Wave 1 projection.
type MissionStatus string

const (
	MissionStatusUnknown   MissionStatus = "unknown"
	MissionStatusPending   MissionStatus = "pending"
	MissionStatusRunning   MissionStatus = "running"
	MissionStatusBlocked   MissionStatus = "blocked"
	MissionStatusDone      MissionStatus = "done"
	MissionStatusFailed    MissionStatus = "failed"
	MissionStatusCancelled MissionStatus = "cancelled"
)

// DecisionCategory identifies the live decision type.
type DecisionCategory string

const (
	DecisionCategoryUnknown  DecisionCategory = "unknown"
	DecisionCategoryApproval DecisionCategory = "approval"
)

// MissionControlSnapshot is the deterministic projection consumed by Mission Control.
type MissionControlSnapshot struct {
	Header                  HeaderView
	Missions                []MissionView
	Decision                *DecisionView
	Activities              []ActivityView
	HiddenMissionCount      int
	HiddenActivityCount     int
	MissionOverflowSummary  string
	ActivityOverflowSummary string
	Degraded                bool
	GeneratedAt             time.Time
}

// MissionView is one projected mission row.
type MissionView struct {
	ID            string
	Kind          MissionKind
	Status        MissionStatus
	Title         string
	Detail        string
	NextAction    string
	UpdatedAt     time.Time
	OwnerAgent    string
	RuntimeHint   string
	BlockedReason string
}

// DecisionView is the single live decision row for Wave 1.
type DecisionView struct {
	ID                   string
	Category             DecisionCategory
	Title                string
	Reason               string
	EffectText           string
	RiskLevel            string
	RiskLabel            string
	ApproveLabel         string
	DenyLabel            string
	AllowForSessionLabel string
	UpdatedAt            time.Time
}

// ActivityView is one recent deterministic timeline row.
type ActivityView struct {
	Kind      MissionActivityKind
	Summary   string
	Timestamp time.Time
}

// HeaderView is the compact Mission Control header summary.
type HeaderView struct {
	ActiveAgentSummary   string
	ModelProviderSummary string
	PendingDecisionCount int
	DegradedNote         string
	ContextSummary       string
	MetricsSummary       string
}
