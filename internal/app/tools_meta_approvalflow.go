package app

import (
	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/exportability"
)

type artifactReleaseApprovalReceipt struct {
	ArtifactLabel      string  `json:"artifact_label"`
	RequestedScope     string  `json:"requested_scope"`
	ExportabilityState string  `json:"exportability_state"`
	OverrideRequested  bool    `json:"override_requested"`
	HighRisk           bool    `json:"high_risk"`
	Decision           string  `json:"decision"`
	Reason             string  `json:"reason"`
	Issue              string  `json:"issue"`
	Fulfillment        string  `json:"fulfillment"`
	FulfillmentRatio   float64 `json:"fulfillment_ratio"`
	SettlementHint     string  `json:"settlement_hint"`
}

func newArtifactReleaseApprovalReceipt(
	artifactLabel string,
	requestedScope string,
	state exportability.DecisionState,
	outcome approvalflow.ArtifactReleaseOutcome,
	overrideRequested bool,
	highRisk bool,
) artifactReleaseApprovalReceipt {
	return artifactReleaseApprovalReceipt{
		ArtifactLabel:      artifactLabel,
		RequestedScope:     requestedScope,
		ExportabilityState: string(state),
		OverrideRequested:  overrideRequested,
		HighRisk:           highRisk,
		Decision:           string(outcome.Decision),
		Reason:             outcome.Reason,
		Issue:              string(outcome.Issue),
		Fulfillment:        string(outcome.Fulfillment),
		FulfillmentRatio:   outcome.FulfillmentRatio,
		SettlementHint:     string(outcome.SettlementHint),
	}
}

func evaluateArtifactReleaseApproval(
	artifactLabel string,
	requestedScope string,
	state exportability.DecisionState,
	overrideRequested bool,
	highRisk bool,
) approvalflow.ArtifactReleaseOutcome {
	return approvalflow.EvaluateArtifactRelease(approvalflow.ArtifactReleaseInput{
		ArtifactLabel:          artifactLabel,
		RequestedArtifactLabel: requestedScope,
		Exportability: exportability.Receipt{
			State: state,
		},
		OverrideRequested: overrideRequested,
		HighRisk:          highRisk,
	})
}

func (r artifactReleaseApprovalReceipt) Details() map[string]interface{} {
	return map[string]interface{}{
		"artifact_label":      r.ArtifactLabel,
		"requested_scope":     r.RequestedScope,
		"exportability_state": r.ExportabilityState,
		"override_requested":  r.OverrideRequested,
		"high_risk":           r.HighRisk,
		"decision":            r.Decision,
		"reason":              r.Reason,
		"issue":               r.Issue,
		"fulfillment":         r.Fulfillment,
		"fulfillment_ratio":   r.FulfillmentRatio,
		"settlement_hint":     r.SettlementHint,
	}
}
