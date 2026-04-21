package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/paymentapproval"
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
	_, hasSubmissionReceiptID := props["submission_receipt_id"]
	_, hasAmount := props["amount"]
	_, hasTrustScore := props["trust_score"]
	_, hasUserMaxPrepay := props["user_max_prepay"]
	_, hasRemainingBudget := props["remaining_budget"]
	_, hasEscrowBuyerDID := props["escrow_buyer_did"]
	_, hasEscrowSellerDID := props["escrow_seller_did"]
	_, hasEscrowReason := props["escrow_reason"]
	_, hasEscrowTaskID := props["escrow_task_id"]
	_, hasEscrowMilestones := props["escrow_milestones"]
	assert.True(t, hasTransactionReceiptID)
	assert.True(t, hasSubmissionReceiptID)
	assert.True(t, hasAmount)
	assert.True(t, hasTrustScore)
	assert.True(t, hasUserMaxPrepay)
	assert.True(t, hasRemainingBudget)
	assert.True(t, hasEscrowBuyerDID)
	assert.True(t, hasEscrowSellerDID)
	assert.True(t, hasEscrowReason)
	assert.True(t, hasEscrowTaskID)
	assert.True(t, hasEscrowMilestones)

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "transaction_receipt_id")
	assert.Contains(t, required, "submission_receipt_id")
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
		"submission_receipt_id":  submission.SubmissionReceiptID,
		"amount":                 "2.00",
		"trust_score":            0.95,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "9.00",
	})
	require.NoError(t, err)

	payload := got.(upfrontPaymentApprovalReceipt)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, submission.SubmissionReceiptID, payload.SubmissionReceiptID)
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
	assert.Empty(t, payload.EscrowExecutionStatus)

	gotSubmission, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, submission.SubmissionReceiptID, gotSubmission.SubmissionReceiptID)
	require.Len(t, events, 1)
	assert.Equal(t, receipts.EventPaymentApproval, events[0].Type)
	assert.Equal(t, submission.SubmissionReceiptID, events[0].SubmissionReceiptID)
	assert.Equal(t, "approval", events[0].Source)
	assert.Equal(t, "approval.upfront_payment", events[0].Subtype)
}

func TestApproveUpfrontPayment_BindsEscrowExecutionInputWhenModeEscrow(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-upfront-escrow",
		ArtifactLabel:       "artifact/upfront-escrow",
		PayloadHash:         "hash-upfront-escrow",
		SourceLineageDigest: "lineage-upfront-escrow",
	})
	require.NoError(t, err)

	got, err := approveUpfrontPayment(ctx, store, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  submission.SubmissionReceiptID,
		"amount":                 "25.00",
		"trust_score":            0.30,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "50.00",
		"escrow_buyer_did":       "did:lango:buyer",
		"escrow_seller_did":      "did:lango:seller",
		"escrow_reason":          "knowledge exchange",
		"escrow_task_id":         "task-upfront-escrow",
		"escrow_milestones": []interface{}{
			map[string]interface{}{"description": "draft", "amount": "10.00"},
			map[string]interface{}{"description": "final", "amount": "15.00"},
		},
	}, func(paymentapproval.Input) paymentapproval.Outcome {
		return paymentapproval.Outcome{
			Decision:      paymentapproval.DecisionApprove,
			Reason:        "Upfront payment approved.",
			SuggestedMode: paymentapproval.ModeEscrow,
			AmountClass:   paymentapproval.AmountMedium,
			RiskClass:     paymentapproval.RiskLow,
		}
	})
	require.NoError(t, err)

	payload := got.(upfrontPaymentApprovalReceipt)
	assert.Equal(t, "escrow", payload.SuggestedMode)
	assert.Equal(t, string(receipts.EscrowExecutionStatusPending), payload.EscrowExecutionStatus)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.NotNil(t, updatedTx.EscrowExecutionInput)
	assert.Equal(t, "did:lango:buyer", updatedTx.EscrowExecutionInput.BuyerDID)
	assert.Equal(t, "did:lango:seller", updatedTx.EscrowExecutionInput.SellerDID)
	assert.Equal(t, "25.00", updatedTx.EscrowExecutionInput.Amount)
	assert.Equal(t, "knowledge exchange", updatedTx.EscrowExecutionInput.Reason)
	assert.Equal(t, "task-upfront-escrow", updatedTx.EscrowExecutionInput.TaskID)
	require.Len(t, updatedTx.EscrowExecutionInput.Milestones, 2)
	assert.Equal(t, "draft", updatedTx.EscrowExecutionInput.Milestones[0].Description)
	assert.Equal(t, "10.00", updatedTx.EscrowExecutionInput.Milestones[0].Amount)
	assert.Equal(t, "final", updatedTx.EscrowExecutionInput.Milestones[1].Description)
	assert.Equal(t, "15.00", updatedTx.EscrowExecutionInput.Milestones[1].Amount)
	assert.Equal(t, receipts.EscrowExecutionStatusPending, updatedTx.EscrowExecutionStatus)
}

func TestApproveUpfrontPayment_RejectsMalformedEscrowInputWithoutMutatingReceipt(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-upfront-escrow-invalid",
		ArtifactLabel:       "artifact/upfront-escrow-invalid",
		PayloadHash:         "hash-upfront-escrow-invalid",
		SourceLineageDigest: "lineage-upfront-escrow-invalid",
	})
	require.NoError(t, err)

	beforeTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipts.PaymentApprovalPending, beforeTx.CurrentPaymentApprovalStatus)
	require.Empty(t, beforeTx.CanonicalSettlementHint)
	require.Empty(t, beforeTx.CanonicalDecision)

	_, err = approveUpfrontPayment(ctx, store, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  submission.SubmissionReceiptID,
		"amount":                 "25.00",
		"trust_score":            0.30,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "50.00",
		"escrow_buyer_did":       "did:lango:buyer",
		"escrow_seller_did":      "did:lango:seller",
		"escrow_reason":          "knowledge exchange",
		"escrow_task_id":         "task-upfront-escrow",
		"escrow_milestones": []interface{}{
			map[string]interface{}{"description": "draft"},
		},
	}, func(paymentapproval.Input) paymentapproval.Outcome {
		return paymentapproval.Outcome{
			Decision:      paymentapproval.DecisionApprove,
			Reason:        "Upfront payment approved.",
			SuggestedMode: paymentapproval.ModeEscrow,
			AmountClass:   paymentapproval.AmountMedium,
			RiskClass:     paymentapproval.RiskLow,
		}
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "escrow_milestones[0]")

	afterTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.PaymentApprovalPending, afterTx.CurrentPaymentApprovalStatus)
	assert.Empty(t, afterTx.CanonicalSettlementHint)
	assert.Empty(t, afterTx.CanonicalDecision)
	assert.Nil(t, afterTx.EscrowExecutionInput)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestApproveUpfrontPayment_ReportsMissingSubmissionReceiptID(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	_, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-upfront-missing-submission",
		ArtifactLabel:       "artifact/upfront",
		PayloadHash:         "hash-upfront",
		SourceLineageDigest: "lineage-upfront",
	})
	require.NoError(t, err)

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "approve_upfront_payment")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"amount":                 "2.00",
		"trust_score":            0.95,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "9.00",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submission_receipt_id")
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
