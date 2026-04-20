package approvalflow

func ApproveArtifactRelease(in ArtifactReleaseInput) ArtifactReleaseOutcome {
	out := ArtifactReleaseOutcome{
		Object:         ObjectArtifactRelease,
		SettlementHint: SettlementHold,
	}

	if in.Exportability.State == "needs-human-review" || in.HighRisk {
		out.Decision = DecisionEscalate
		out.Issue = IssuePolicy
		out.Reason = "Artifact release requires human escalation."
		out.SettlementHint = SettlementReview
		return out
	}

	if in.Exportability.State == "blocked" {
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

	if in.ArtifactLabel != in.RequestedScope {
		out.Decision = DecisionRequestRevision
		out.Issue = IssueScopeMismatch
		out.Fulfillment = FulfillmentPartial
		out.Reason = "Submitted artifact does not match requested scope."
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
