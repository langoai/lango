package exportability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluate_PublicAndUserExportableSourcesAllowExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "pub-1", AssetLabel: "docs/api", Class: ClassPublic},
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
	})

	assert.Equal(t, StageFinal, receipt.Stage)
	assert.Equal(t, StateExportable, receipt.State)
	assert.Equal(t, "allowed_user_exportable", receipt.PolicyCode)
	require.Len(t, receipt.Lineage, 2)
	assert.Equal(t, "pub-1", receipt.Lineage[0].AssetID)
	assert.Equal(t, ClassPublic, receipt.Lineage[0].Class)
}

func TestEvaluate_PrivateSourceBlocksExport(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "usr-1", AssetLabel: "user/wiki", Class: ClassUserExportable},
		{AssetID: "priv-1", AssetLabel: "private/chat", Class: ClassPrivateConfidential},
	})

	assert.Equal(t, StageFinal, receipt.Stage)
	assert.Equal(t, StateBlocked, receipt.State)
	assert.Equal(t, "blocked_private_source", receipt.PolicyCode)
	require.Len(t, receipt.Lineage, 2)
	assert.Equal(t, "highest_sensitivity_wins", receipt.Lineage[1].Rule)
}

func TestEvaluate_MissingSourceMetadataRequiresHumanReview(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageDraft, []SourceRef{
		{AssetID: "unknown-1", AssetLabel: "mystery", Class: ""},
	})

	assert.Equal(t, StageDraft, receipt.Stage)
	assert.Equal(t, StateNeedsHumanReview, receipt.State)
	assert.Equal(t, "review_metadata_conflict", receipt.PolicyCode)
	require.Len(t, receipt.Lineage, 1)
	assert.Equal(t, "metadata_missing", receipt.Lineage[0].Rule)
}

func TestEvaluate_DisabledPolicyPreservesLineage(t *testing.T) {
	policy := Policy{Enabled: false}
	receipt := Evaluate(policy, StageDraft, []SourceRef{
		{AssetID: "pub-1", AssetLabel: "docs/api", Class: ClassPublic},
	})

	assert.Equal(t, StageDraft, receipt.Stage)
	assert.Equal(t, StateNeedsHumanReview, receipt.State)
	assert.Equal(t, "review_policy_disabled", receipt.PolicyCode)
	require.Len(t, receipt.Lineage, 1)
	assert.Equal(t, "pub-1", receipt.Lineage[0].AssetID)
	assert.Equal(t, ClassPublic, receipt.Lineage[0].Class)
}

func TestEvaluate_UnknownSourceClassRequiresHumanReview(t *testing.T) {
	policy := Policy{Enabled: true}
	receipt := Evaluate(policy, StageFinal, []SourceRef{
		{AssetID: "mystery-1", AssetLabel: "mystery", Class: SourceClass("partner-private")},
	})

	assert.Equal(t, StageFinal, receipt.Stage)
	assert.Equal(t, StateNeedsHumanReview, receipt.State)
	assert.Equal(t, "review_metadata_conflict", receipt.PolicyCode)
	require.Len(t, receipt.Lineage, 1)
	assert.Equal(t, "metadata_invalid", receipt.Lineage[0].Rule)
}
