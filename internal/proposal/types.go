package proposal

import "time"

type ProposalStatus string

const (
	ProposalStatusSuggested ProposalStatus = "suggested"
	ProposalStatusPreparing ProposalStatus = "preparing"
	ProposalStatusPrepared  ProposalStatus = "prepared"
	ProposalStatusDismissed ProposalStatus = "dismissed"
	ProposalStatusAccepted  ProposalStatus = "accepted"
	ProposalStatusExpired   ProposalStatus = "expired"
)

type ProposalSource struct {
	Kind string
	Ref  string
}

type PreparedBrief struct {
	SourceSummary             string
	Reason                    string
	SuggestedAcceptanceEffect string
	SupportingEvidence        []string
}

type Proposal struct {
	ProposalID    string
	SessionKey    string
	Source        ProposalSource
	Title         string
	Summary       string
	Reason        string
	Confidence    float64
	Status        ProposalStatus
	PreparedBrief *PreparedBrief
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
}

type UpsertInput struct {
	SessionKey string
	Source     ProposalSource
	Title      string
	Summary    string
	Reason     string
	Confidence float64
	ExpiresAt  time.Time
}
