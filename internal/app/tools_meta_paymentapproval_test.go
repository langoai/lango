package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaTools_IncludesApproveUpfrontPayment(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	tool := findTool(tools, "approve_upfront_payment")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	_, hasAmount := props["amount"]
	_, hasTrustScore := props["trust_score"]
	_, hasUserMaxPrepay := props["user_max_prepay"]
	_, hasRemainingBudget := props["remaining_budget"]
	assert.True(t, hasTransactionReceiptID)
	assert.True(t, hasAmount)
	assert.True(t, hasTrustScore)
	assert.True(t, hasUserMaxPrepay)
	assert.True(t, hasRemainingBudget)

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "transaction_receipt_id")
	assert.Contains(t, required, "amount")
	assert.Contains(t, required, "trust_score")
	assert.Contains(t, required, "user_max_prepay")
	assert.Contains(t, required, "remaining_budget")
}

func TestApproveUpfrontPayment_UpdatesTransactionAndReturnsDecisionPayload(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-upfront",
		ArtifactLabel:       "artifact/upfront",
		PayloadHash:         "hash-upfront",
		SourceLineageDigest: "lineage-upfront",
	})
	require.NoError(t, err)

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "approve_upfront_payment")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"amount":                 "2.00",
		"trust_score":            0.95,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "9.00",
	})
	require.NoError(t, err)

	payload := got.(upfrontPaymentApprovalReceipt)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, "2.00", payload.Amount)
	assert.Equal(t, 0.95, payload.TrustScore)
	assert.Equal(t, "5.00", payload.UserMaxPrepay)
	assert.Equal(t, "9.00", payload.RemainingBudget)
	assert.Equal(t, "approve", payload.Decision)
	assert.Equal(t, "Upfront payment approved.", payload.Reason)
	assert.Equal(t, "prepay", payload.SuggestedMode)
	assert.Equal(t, "low", payload.AmountClass)
	assert.Equal(t, "low", payload.RiskClass)
	assert.Equal(t, string(receipts.PaymentApprovalApproved), payload.CurrentPaymentApprovalStatus)
	assert.Equal(t, "approve", payload.CanonicalDecision)
	assert.Equal(t, "prepay", payload.CanonicalSettlementHint)

	gotSubmission, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, submission.SubmissionReceiptID, gotSubmission.SubmissionReceiptID)
	require.Len(t, events, 1)
	assert.Equal(t, receipts.EventPaymentApproval, events[0].Type)
	assert.Equal(t, submission.SubmissionReceiptID, events[0].SubmissionReceiptID)
	assert.Equal(t, "approval", events[0].Source)
	assert.Equal(t, "approval.upfront_payment", events[0].Subtype)
}

func TestApproveUpfrontPayment_ReportsMissingReceiptsDependency(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "approve_upfront_payment")
	require.NotNil(t, tool)

	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"transaction_receipt_id": "tx-missing",
		"amount":                 "2.00",
		"trust_score":            0.95,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "9.00",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "receipts store dependency is not configured")
}
