package approvalflow

import (
	"testing"

	"github.com/langoai/lango/internal/exportability"
	"github.com/stretchr/testify/assert"
)

func TestApproveArtifactRelease_ApproveOnExportableScopeMatch(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:  "research memo",
		RequestedScope: "research memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionApprove, outcome.Decision)
	assert.Equal(t, SettlementAutoRelease, outcome.SettlementHint)
}

func TestApproveArtifactRelease_RequestRevisionOnScopeMismatch(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:  "rough draft",
		RequestedScope: "final design memo",
		Exportability: exportability.Receipt{
			State: exportability.StateExportable,
		},
	})

	assert.Equal(t, DecisionRequestRevision, outcome.Decision)
	assert.Equal(t, IssueScopeMismatch, outcome.Issue)
}

func TestApproveArtifactRelease_EscalateOnNeedsHumanReview(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:  "sensitive memo",
		RequestedScope: "sensitive memo",
		Exportability: exportability.Receipt{
			State: exportability.StateNeedsHumanReview,
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}

func TestApproveArtifactRelease_RejectOnBlockedOverrideAttempt(t *testing.T) {
	outcome := ApproveArtifactRelease(ArtifactReleaseInput{
		ArtifactLabel:  "blocked memo",
		RequestedScope: "blocked memo",
		Exportability: exportability.Receipt{
			State: exportability.StateBlocked,
		},
		OverrideRequested: false,
	})

	assert.Equal(t, DecisionReject, outcome.Decision)
	assert.Equal(t, IssuePolicy, outcome.Issue)
}
