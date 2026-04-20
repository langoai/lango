package approvalflow

import "github.com/langoai/lango/internal/exportability"

func EvaluateArtifactRelease(in ArtifactReleaseInput) ArtifactReleaseOutcome {
	out := ArtifactReleaseOutcome{
		Object:         ObjectArtifactRelease,
		SettlementHint: SettlementHold,
	}

	if in.Exportability.State == exportability.StateNeedsHumanReview || in.HighRisk {
		out.Decision = DecisionEscalate
		out.Issue = IssuePolicy
		out.Reason = "Artifact release requires human escalation."
		out.SettlementHint = SettlementReview
		return out
	}

	if in.Exportability.State == exportability.StateBlocked {
		if in.OverrideRequested {
			out.Decision = DecisionEscalate
			out.Issue = IssuePolicy
			out.Reason = "Blocked artifact override requires human approval."
			out.SettlementHint = SettlementReview
			return out
		}

		out.Decision = DecisionReject
		out.Issue = IssuePolicy
		out.Fulfillment = FulfillmentNone
		out.Reason = "Artifact release blocked by exportability policy."
		return out
	}

	if in.Exportability.State != exportability.StateExportable {
		out.Decision = DecisionRequestRevision
		out.Issue = IssuePolicy
		out.Fulfillment = FulfillmentPartial
		out.Reason = "Artifact release requires exportable receipt state."
		out.SettlementHint = SettlementReview
		return out
	}

	if in.ArtifactLabel != in.RequestedArtifactLabel {
		out.Decision = DecisionRequestRevision
		out.Issue = IssueScopeMismatch
		out.Fulfillment = FulfillmentPartial
		out.Reason = "Submitted artifact label does not match requested artifact label."
		out.SettlementHint = SettlementReview
		return out
	}

	out.Decision = DecisionApprove
	out.Reason = "Artifact release approved."
	out.Fulfillment = FulfillmentSubstantial
	out.FulfillmentRatio = 1.0
	out.SettlementHint = SettlementAutoRelease
	return out
}
