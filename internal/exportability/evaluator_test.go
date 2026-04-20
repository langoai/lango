package exportability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluate_PublicAndUserExportableSourcesAllowExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "pub-1", AssetLabel: "docs/api", Class: ClassPublic},
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
	})

	assert.Equal(t, StateExportable, receipt.State)
	assert.Equal(t, "allowed_user_exportable", receipt.PolicyCode)
}

func TestEvaluate_PrivateSourceBlocksExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
		{AssetID: "priv-1", AssetLabel: "private/chat", Class: ClassPrivateConfidential},
	})

	assert.Equal(t, StateBlocked, receipt.State)
	assert.Equal(t, "blocked_private_source", receipt.PolicyCode)
}

func TestEvaluate_MissingSourceMetadataRequiresHumanReview(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageDraft, []SourceRef{
		{AssetID: "unknown-1", AssetLabel: "mystery", Class: ""},
	})

	assert.Equal(t, StateNeedsHumanReview, receipt.State)
	assert.Equal(t, "review_metadata_conflict", receipt.PolicyCode)
}
