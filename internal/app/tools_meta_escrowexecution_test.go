package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaToolsWithEscrow_IncludesExecuteEscrowRecommendation(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	escrowEngine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())

	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, store, escrowEngine)
	tool := findTool(tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.SafetyLevelDangerous, tool.SafetyLevel)
}

func TestBuildMetaToolsWithEscrow_ExecuteEscrowRecommendationRequiresTransactionReceiptIDOnly(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	escrowEngine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())

	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, store, escrowEngine)
	tool := findTool(tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)

	props, ok := tool.Parameters["properties"].(map[string]interface{})
	require.True(t, ok)
	require.Len(t, props, 1)
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	assert.True(t, hasTransactionReceiptID)

	required, ok := tool.Parameters["required"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"transaction_receipt_id"}, required)
}

func TestExecuteEscrowRecommendation_SuccessfulPayloadOnPreparedEscrowRecommendedTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-execute-escrow",
		ArtifactLabel:       "artifact/execute-escrow",
		PayloadHash:         "hash-execute-escrow",
		SourceLineageDigest: "lineage-execute-escrow",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
		AmountClass:   paymentapproval.AmountMedium,
		RiskClass:     paymentapproval.RiskLow,
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "25.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-execute-escrow",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "draft", Amount: "10.00"},
			{Description: "final", Amount: "15.00"},
		},
	})
	require.NoError(t, err)

	escrowEngine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, store, escrowEngine)
	tool := findTool(tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload["transaction_receipt_id"])
	assert.Equal(t, submission.SubmissionReceiptID, payload["submission_receipt_id"])
	assert.NotEmpty(t, payload["escrow_reference"])
	assert.Equal(t, string(receipts.EscrowExecutionStatusFunded), payload["escrow_execution_status"])

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.EscrowExecutionStatusFunded, updatedTx.EscrowExecutionStatus)
	assert.NotEmpty(t, updatedTx.EscrowReference)
}
