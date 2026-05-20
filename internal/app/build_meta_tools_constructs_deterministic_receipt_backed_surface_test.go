package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	"github.com/langoai/lango/internal/exportability"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaTools_ConstructsDeterministicReceiptBackedSurface(t *testing.T) {
	t.Parallel()

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), receipts.NewStore())
	repeated := buildMetaTools(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), receipts.NewStore())
	assert.Equal(t, buildMetaToolsConstructsDeterministicReceiptBackedSurfaceToolNames(tools), buildMetaToolsConstructsDeterministicReceiptBackedSurfaceToolNames(repeated))

	for _, name := range []string{
		"save_knowledge",
		"evaluate_exportability",
		"approve_artifact_release",
		"create_dispute_ready_receipt",
		"open_knowledge_exchange_transaction",
		"select_knowledge_exchange_path",
		"approve_upfront_payment",
		"apply_settlement_progression",
		"get_knowledge_history",
		"search_knowledge",
		"save_learning",
		"search_learnings",
		"create_skill",
		"list_skills",
		"view_skill",
		"import_skill",
		"learning_stats",
		"learning_cleanup",
		"list_dead_lettered_post_adjudication_executions",
		"get_post_adjudication_execution_status",
		"retry_post_adjudication_execution",
	} {
		tool := findTool(tools, name)
		require.NotNil(t, tool, "expected %s to be registered", name)
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.Handler)
		assert.Equal(t, "object", tool.Parameters["type"])
	}

	for _, name := range []string{
		"execute_settlement",
		"execute_partial_settlement",
		"execute_escrow_recommendation",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
	} {
		assert.Nil(t, findTool(tools, name), "expected %s to require an explicit runtime", name)
	}
	assertNoDuplicateNames(t, tools)
}

func buildMetaToolsConstructsDeterministicReceiptBackedSurfaceToolNames(tools []*agent.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}

func TestEvaluateExportability_RejectsUnavailableStoreBeforeParams(t *testing.T) {
	t.Parallel()

	tool := findTool(
		buildMetaTools(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), receipts.NewStore()),
		"evaluate_exportability",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "knowledge store is not available")
}

func TestExportabilityHelpers_DefaultPolicyAndSourceRefBranches(t *testing.T) {
	t.Parallel()

	assert.True(t, exportabilityPolicyEnabled(nil))

	disabled := config.DefaultConfig()
	disabled.Security.Exportability.Enabled = false
	assert.False(t, exportabilityPolicyEnabled(disabled))

	refs, err := exportabilitySourceRefs([]knowledge.KnowledgeEntry{
		{
			Key:         "source-with-label-fallback",
			Category:    entknowledge.CategoryFact,
			Content:     "private source",
			SourceClass: "",
		},
		{
			Key:         "source-with-label",
			Category:    entknowledge.CategoryFact,
			Content:     "public source",
			SourceClass: string(exportability.ClassPublic),
			AssetLabel:  "artifact/source-label",
		},
	})
	require.NoError(t, err)
	require.Len(t, refs, 2)
	assert.Equal(t, "source-with-label-fallback", refs[0].AssetLabel)
	assert.Equal(t, exportability.ClassPrivateConfidential, refs[0].Class)
	assert.Equal(t, "artifact/source-label", refs[1].AssetLabel)
	assert.Equal(t, exportability.ClassPublic, refs[1].Class)

	_, err = exportabilitySourceRefs([]knowledge.KnowledgeEntry{
		{
			Key:         "bad-source-class",
			Category:    entknowledge.CategoryFact,
			Content:     "bad source",
			SourceClass: "partner-only",
		},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, `parse source class for "bad-source-class"`)
}

func TestExportabilityReceiptPayload_MapsLineageRows(t *testing.T) {
	t.Parallel()

	payload := exportabilityReceiptPayload("artifact/buildMetaToolsConstructsDeterministicReceiptBackedSurface4", exportability.Receipt{
		Stage:       exportability.StageFinal,
		State:       exportability.StateExportable,
		PolicyCode:  "allowed_public_only",
		Explanation: "Artifact is exportable.",
		Lineage: []exportability.LineageSummary{
			{
				AssetID:    "source-1",
				AssetLabel: "artifact/source-1",
				Class:      exportability.ClassPublic,
				Rule:       "source_class_ok",
			},
		},
	})

	assert.Equal(t, "artifact/buildMetaToolsConstructsDeterministicReceiptBackedSurface4", payload["artifact_label"])
	assert.Equal(t, "final", payload["stage"])
	assert.Equal(t, "exportable", payload["state"])
	assert.Equal(t, "allowed_public_only", payload["policy_code"])
	assert.Equal(t, "Artifact is exportable.", payload["explanation"])

	lineage, ok := payload["lineage"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, lineage, 1)
	assert.Equal(t, "source-1", lineage[0]["asset_id"])
	assert.Equal(t, "artifact/source-1", lineage[0]["asset_label"])
	assert.Equal(t, "public", lineage[0]["class"])
	assert.Equal(t, "source_class_ok", lineage[0]["rule"])
}

func TestApproveUpfrontPayment_CapturesEvaluatorInputAndEscalatesWithoutEscrowFields(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-buildMetaToolsConstructsDeterministicReceiptBackedSurface4-payment",
		ArtifactLabel:       "artifact/buildMetaToolsConstructsDeterministicReceiptBackedSurface4-payment",
		PayloadHash:         "hash-buildMetaToolsConstructsDeterministicReceiptBackedSurface4-payment",
		SourceLineageDigest: "lineage-buildMetaToolsConstructsDeterministicReceiptBackedSurface4-payment",
	})
	require.NoError(t, err)

	var captured paymentapproval.Input
	got, err := approveUpfrontPayment(ctx, store, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  submission.SubmissionReceiptID,
		"amount":                 "125.00",
		"trust_score":            0.20,
		"user_max_prepay":        "200.00",
		"remaining_budget":       "300.00",
	}, func(in paymentapproval.Input) paymentapproval.Outcome {
		captured = in
		return paymentapproval.Outcome{
			Decision:      paymentapproval.DecisionEscalate,
			Reason:        "Manual review required.",
			PolicyCode:    "escalate_policy",
			SuggestedMode: paymentapproval.ModeEscalate,
			AmountClass:   paymentapproval.AmountCritical,
			RiskClass:     paymentapproval.RiskHigh,
		}
	})
	require.NoError(t, err)

	assert.Equal(t, "125.00", captured.Amount)
	assert.Equal(t, 0.20, captured.Trust.Score)
	assert.Equal(t, "200.00", captured.Budget.UserMaxPrepay)
	assert.Equal(t, "300.00", captured.Budget.RemainingBudget)

	payload, ok := got.(upfrontPaymentApprovalReceipt)
	require.True(t, ok)
	assert.Equal(t, "escalate", payload.Decision)
	assert.Equal(t, "Manual review required.", payload.Reason)
	assert.Equal(t, "escalate", payload.SuggestedMode)
	assert.Equal(t, "critical", payload.AmountClass)
	assert.Equal(t, "high", payload.RiskClass)
	assert.Equal(t, string(receipts.PaymentApprovalEscalated), payload.CurrentPaymentApprovalStatus)
	assert.Equal(t, "escalate", payload.CanonicalDecision)
	assert.Equal(t, "escalate", payload.CanonicalSettlementHint)
	assert.Empty(t, payload.EscrowExecutionStatus)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Nil(t, updatedTx.EscrowExecutionInput)
	assert.Empty(t, updatedTx.EscrowExecutionStatus)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, receipts.EventPaymentApproval, events[0].Type)
	assert.Equal(t, "approval.upfront_payment", events[0].Subtype)
}

func TestParseEscrowExecutionInput_RejectsMalformedMilestoneShapes(t *testing.T) {
	t.Parallel()

	baseParams := map[string]interface{}{
		"escrow_buyer_did":  "did:lango:buyer",
		"escrow_seller_did": "did:lango:seller",
		"escrow_reason":     "knowledge exchange",
	}

	tests := []struct {
		name      string
		giveRaw   interface{}
		wantError string
	}{
		{
			name:      "milestones not array",
			giveRaw:   "draft:1.00",
			wantError: "escrow_milestones must be an array",
		},
		{
			name:      "milestone not object",
			giveRaw:   []interface{}{"draft"},
			wantError: "escrow_milestones[0] must be an object",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := make(map[string]interface{}, len(baseParams)+1)
			for key, value := range baseParams {
				params[key] = value
			}
			params["escrow_milestones"] = tt.giveRaw

			got, err := parseEscrowExecutionInput(params, "1.00")
			require.Error(t, err)
			assert.Equal(t, receipts.EscrowExecutionInput{}, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}
