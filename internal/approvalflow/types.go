package approvalflow

import "github.com/langoai/lango/internal/exportability"

type ApprovalObject string

const (
	ObjectUpfrontPayment  ApprovalObject = "upfront_payment"
	ObjectArtifactRelease ApprovalObject = "artifact_release"
)

type Decision string

const (
	DecisionApprove         Decision = "approve"
	DecisionReject          Decision = "reject"
	DecisionRequestRevision Decision = "request-revision"
	DecisionEscalate        Decision = "escalate"
)

type IssueClass string

const (
	IssueScopeMismatch IssueClass = "scope_mismatch"
	IssueQuality       IssueClass = "quality_issue"
	IssuePolicy        IssueClass = "policy_issue"
)

type FulfillmentGrade string

const (
	FulfillmentNone        FulfillmentGrade = "none"
	FulfillmentPartial     FulfillmentGrade = "partial"
	FulfillmentSubstantial FulfillmentGrade = "substantial"
)

type SettlementHint string

const (
	SettlementAutoRelease SettlementHint = "auto_release"
	SettlementHold        SettlementHint = "hold"
	SettlementReview      SettlementHint = "review"
)

type ArtifactReleaseInput struct {
	ArtifactLabel     string
	RequestedScope    string
	Exportability     exportability.Receipt
	OverrideRequested bool
	HighRisk          bool
}

type ArtifactReleaseOutcome struct {
	Object           ApprovalObject   `json:"object"`
	Decision         Decision         `json:"decision"`
	Reason           string           `json:"reason"`
	Issue            IssueClass       `json:"issue,omitempty"`
	Fulfillment      FulfillmentGrade `json:"fulfillment,omitempty"`
	FulfillmentRatio float64          `json:"fulfillment_ratio,omitempty"`
	SettlementHint   SettlementHint   `json:"settlement_hint"`
}
