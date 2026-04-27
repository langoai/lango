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

func createSubmittedTransaction(t *testing.T, store *receipts.Store, ctx context.Context, transactionID string) receipts.TransactionReceipt {
	t.Helper()

	_, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  transactionID,
		Counterparty:   "did:lango:peer-" + transactionID,
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	_, transaction, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       transactionID,
		ArtifactLabel:       "artifact-" + transactionID,
		PayloadHash:         "hash-" + transactionID,
		SourceLineageDigest: "lineage-" + transactionID,
	})
	require.NoError(t, err)

	return transaction
}

func TestBuildMetaTools_IncludesApplySettlementProgression(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	tool := findTool(tools, "apply_settlement_progression")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	_, hasOutcome := props["outcome"]
	_, hasReason := props["reason"]
	_, hasPartialHint := props["partial_hint"]
	assert.True(t, hasTransactionReceiptID)
	assert.True(t, hasOutcome)
	assert.True(t, hasReason)
	assert.True(t, hasPartialHint)

	outcomeSchema, ok := props["outcome"].(map[string]interface{})
	require.True(t, ok)
	enum, ok := outcomeSchema["enum"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"approve", "reject", "request-revision", "escalate"}, enum)

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"transaction_receipt_id", "outcome"}, required)
}

func TestApplySettlementProgression_ApprovePathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "apply_settlement_progression")
	require.NotNil(t, tool)

	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-settlement-progress-approve")

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "approve",
	})
	require.NoError(t, err)

	payload, ok := got.(applySettlementProgressionReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionApprovedForSettlement), payload.SettlementProgressionStatus)
	assert.Equal(t, string(receipts.SettlementProgressionReasonCodeApprove), payload.SettlementProgressionReasonCode)
	assert.Equal(t, "Artifact release approved.", payload.SettlementProgressionReason)
	assert.Empty(t, payload.PartialHint)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.SettlementProgressionApprovedForSettlement, updatedTx.SettlementProgressionStatus)
	assert.Equal(t, receipts.SettlementProgressionReasonCodeApprove, updatedTx.SettlementProgressionReasonCode)
	assert.Equal(t, "Artifact release approved.", updatedTx.SettlementProgressionReason)
}

func TestApplySettlementProgression_PreservesReasonAndPartialHint(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "apply_settlement_progression")
	require.NotNil(t, tool)

	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-settlement-progress-revision")

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "request-revision",
		"reason":                 "Need a revised final draft.",
		"partial_hint":           "partial delivery is acceptable",
	})
	require.NoError(t, err)

	payload, ok := got.(applySettlementProgressionReceipt)
	require.True(t, ok)
	assert.Equal(t, string(receipts.SettlementProgressionReviewNeeded), payload.SettlementProgressionStatus)
	assert.Equal(t, string(receipts.SettlementProgressionReasonCodeRequestRevision), payload.SettlementProgressionReasonCode)
	assert.Equal(t, "Need a revised final draft.", payload.SettlementProgressionReason)
	assert.Equal(t, "partial delivery is acceptable", payload.PartialHint)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.SettlementProgressionReviewNeeded, updatedTx.SettlementProgressionStatus)
	assert.Equal(t, receipts.SettlementProgressionReasonCodeRequestRevision, updatedTx.SettlementProgressionReasonCode)
	assert.Equal(t, "Need a revised final draft.", updatedTx.SettlementProgressionReason)
	assert.Equal(t, "partial delivery is acceptable", updatedTx.PartialSettlementHint)
}

func TestApplySettlementProgression_EscalateFromReviewNeededReturnsCanonicalDisputeReadyState(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "apply_settlement_progression")
	require.NotNil(t, tool)

	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-settlement-progress-escalate")

	_, err := store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		receipts.SettlementProgressionReviewNeeded,
		receipts.SettlementProgressionReasonCodeReject,
		"Artifact release blocked by policy.",
		"",
	)
	require.NoError(t, err)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "escalate",
		"reason":                 "manual approval required",
	})
	require.NoError(t, err)

	payload, ok := got.(applySettlementProgressionReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionDisputeReady), payload.SettlementProgressionStatus)
	assert.Equal(t, string(receipts.SettlementProgressionReasonCodeEscalate), payload.SettlementProgressionReasonCode)
	assert.Equal(t, "manual approval required", payload.SettlementProgressionReason)
	assert.Empty(t, payload.PartialHint)
	assert.Empty(t, payload.DisputeLifecycleStatus)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.SettlementProgressionDisputeReady, updatedTx.SettlementProgressionStatus)
	assert.Equal(t, receipts.SettlementProgressionReasonCodeEscalate, updatedTx.SettlementProgressionReasonCode)
	assert.Equal(t, "manual approval required", updatedTx.SettlementProgressionReason)
	assert.Empty(t, updatedTx.PartialSettlementHint)
	assert.Empty(t, updatedTx.DisputeLifecycleStatus)
}
