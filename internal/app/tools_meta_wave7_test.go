package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/exportability"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/receipts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportabilityPolicyEnabled_UsesDefaultAndExplicitConfig(t *testing.T) {
	t.Parallel()

	assert.True(t, exportabilityPolicyEnabled(nil))

	cfg := config.DefaultConfig()
	cfg.Security.Exportability.Enabled = false
	assert.False(t, exportabilityPolicyEnabled(cfg))

	cfg.Security.Exportability.Enabled = true
	assert.True(t, exportabilityPolicyEnabled(cfg))
}

func TestExportabilitySourceRefs_MapsLabelsAndRejectsInvalidClass(t *testing.T) {
	t.Parallel()

	refs, err := exportabilitySourceRefs([]knowledge.KnowledgeEntry{
		{
			Key:         "policy-source",
			SourceClass: string(exportability.ClassPublic),
			AssetLabel:  "Policy Source",
		},
		{
			Key:         "private-source",
			SourceClass: string(exportability.ClassPrivateConfidential),
		},
	})
	require.NoError(t, err)
	require.Len(t, refs, 2)
	assert.Equal(t, "policy-source", refs[0].AssetID)
	assert.Equal(t, "Policy Source", refs[0].AssetLabel)
	assert.Equal(t, exportability.ClassPublic, refs[0].Class)
	assert.Equal(t, "private-source", refs[1].AssetLabel, "empty asset labels fall back to key")

	_, err = exportabilitySourceRefs([]knowledge.KnowledgeEntry{
		{Key: "bad-source", SourceClass: "secret"},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, `parse source class for "bad-source"`)
}

func TestExportabilityReceiptPayload_PreservesDecisionAndLineage(t *testing.T) {
	t.Parallel()

	payload := exportabilityReceiptPayload("artifact.tar", exportability.Receipt{
		Stage:       exportability.StageFinal,
		State:       exportability.StateNeedsHumanReview,
		PolicyCode:  "mixed-lineage",
		Explanation: "Human review required.",
		Lineage: []exportability.LineageSummary{
			{
				AssetID:    "source-1",
				AssetLabel: "Source One",
				Class:      exportability.ClassUserExportable,
				Rule:       "source_class_ok",
			},
		},
	})

	assert.Equal(t, "artifact.tar", payload["artifact_label"])
	assert.Equal(t, string(exportability.StageFinal), payload["stage"])
	assert.Equal(t, string(exportability.StateNeedsHumanReview), payload["state"])
	assert.Equal(t, "mixed-lineage", payload["policy_code"])
	assert.Equal(t, "Human review required.", payload["explanation"])
	lineage := payload["lineage"].([]map[string]interface{})
	require.Len(t, lineage, 1)
	assert.Equal(t, "source-1", lineage[0]["asset_id"])
	assert.Equal(t, "Source One", lineage[0]["asset_label"])
	assert.Equal(t, string(exportability.ClassUserExportable), lineage[0]["class"])
	assert.Equal(t, "source_class_ok", lineage[0]["rule"])
}

func TestParseEscrowExecutionInput_ParsesOptionalTaskAndMilestones(t *testing.T) {
	t.Parallel()

	got, err := parseEscrowExecutionInput(map[string]interface{}{
		"escrow_buyer_did":  "did:lango:buyer",
		"escrow_seller_did": "did:lango:seller",
		"escrow_reason":     "delivery review",
		"escrow_task_id":    "task-123",
		"escrow_milestones": []interface{}{
			map[string]interface{}{"description": "draft", "amount": "1.00"},
			map[string]interface{}{"description": "final", "amount": "2.00"},
		},
	}, "3.00")
	require.NoError(t, err)

	assert.Equal(t, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "3.00",
		Reason:    "delivery review",
		TaskID:    "task-123",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "draft", Amount: "1.00"},
			{Description: "final", Amount: "2.00"},
		},
	}, got)
}

func TestParseEscrowExecutionInput_RejectsMalformedMilestones(t *testing.T) {
	t.Parallel()

	base := map[string]interface{}{
		"escrow_buyer_did":  "did:lango:buyer",
		"escrow_seller_did": "did:lango:seller",
		"escrow_reason":     "delivery review",
	}

	cases := []struct {
		name      string
		mutate    func(map[string]interface{})
		wantError string
	}{
		{
			name: "milestones not array",
			mutate: func(params map[string]interface{}) {
				params["escrow_milestones"] = "not-array"
			},
			wantError: "escrow_milestones must be an array",
		},
		{
			name: "milestone not object",
			mutate: func(params map[string]interface{}) {
				params["escrow_milestones"] = []interface{}{"not-object"}
			},
			wantError: "escrow_milestones[0] must be an object",
		},
		{
			name: "missing milestone amount",
			mutate: func(params map[string]interface{}) {
				params["escrow_milestones"] = []interface{}{
					map[string]interface{}{"description": "draft"},
				}
			},
			wantError: "escrow_milestones[0]: missing amount parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]interface{}, len(base)+1)
			for k, v := range base {
				params[k] = v
			}
			tt.mutate(params)

			got, err := parseEscrowExecutionInput(params, "1.00")
			require.Error(t, err)
			assert.Equal(t, receipts.EscrowExecutionInput{}, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestPostAdjudicationStatusBackgroundTaskReader_NilDispatcherIsEmpty(t *testing.T) {
	t.Parallel()

	got, err := (postAdjudicationStatusBackgroundTaskReader{}).ListTaskSnapshots(context.Background())
	require.NoError(t, err)
	assert.Nil(t, got)
}
