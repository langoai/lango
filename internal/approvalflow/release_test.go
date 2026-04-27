package approvalflow

import (
	"testing"

	"github.com/langoai/lango/internal/exportability"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateArtifactRelease_ApproveOnExportableArtifactLabelMatch(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "research memo",
		RequestedArtifactLabel: "research memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionApprove, outcome.Decision)
	assert.Equal(t, FulfillmentSubstantial, outcome.Fulfillment)
	assert.Equal(t, 1.0, outcome.FulfillmentRatio)
	assert.Equal(t, SettlementAutoRelease, outcome.SettlementHint)
}

func TestEvaluateArtifactRelease_RequestRevisionOnArtifactLabelMismatch(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "rough draft",
		RequestedArtifactLabel: "final design memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionRequestRevision, outcome.Decision)
	assert.Equal(t, IssueScopeMismatch, outcome.Issue)
}

func TestEvaluateArtifactRelease_EscalateOnNeedsHumanReview(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "sensitive memo",
		RequestedArtifactLabel: "sensitive memo",
		Exportability: exportability.Receipt{
			State: exportability.StateNeedsHumanReview,
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}

func TestEvaluateArtifactRelease_EscalateOnHighRisk(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "sensitive memo",
		RequestedArtifactLabel: "sensitive memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
		HighRisk: true,
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}

func TestEvaluateArtifactRelease_EscalateOnBlockedOverrideRequested(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "blocked memo",
		RequestedArtifactLabel: "blocked memo",
		Exportability: exportability.Receipt{
			State: exportability.StateBlocked,
		},
		OverrideRequested: true,
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}

func TestEvaluateArtifactRelease_RejectOnBlockedWithoutOverride(t *testing.T) {
	outcome := EvaluateArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:          "blocked memo",
		RequestedArtifactLabel: "blocked memo",
		Exportability: exportability.Receipt{
			State: exportability.StateBlocked,
		},
		OverrideRequested: false,
	})

	assert.Equal(t, DecisionReject, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
	assert.Equal(t, FulfillmentNone, outcome.Fulfillment)
}
