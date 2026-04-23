package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/receipts"
)

func replayToolConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Replay.AllowedActors = []string{"operator:alice", "operator:bob"}
	cfg.Replay.ReleaseAllowedActors = []string{"operator:alice"}
	cfg.Replay.RefundAllowedActors = []string{"operator:alice", "operator:bob"}
	return cfg
}

func TestBuildMetaTools_IncludesRetryPostAdjudicationExecution(t *testing.T) {
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, replayToolConfig(), receipts.NewStore(), nil, nil, nil, nil, nil, nil, &fakeAdjudicationBackgroundDispatcher{}), "retry_post_adjudication_execution")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	assert.True(t, hasTransactionReceiptID)

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"transaction_receipt_id"}, required)
}

func TestRetryPostAdjudicationExecution_SuccessReturnsDispatchReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-replay")

	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              receipts.EscrowAdjudicationRelease,
		Reason:               "release adjudicated",
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRelease,
		AttemptCount:         4,
		Reason:               "worker failed repeatedly",
	})
	require.NoError(t, err)

	dispatcher := &fakeAdjudicationBackgroundDispatcher{taskID: "task-replay-123"}
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, replayToolConfig(), store, nil, nil, nil, nil, nil, nil, dispatcher), "retry_post_adjudication_execution")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(retryPostAdjudicationExecutionReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, "escrow-123", payload.EscrowReference)
	assert.Equal(t, "release", payload.Outcome)
	require.NotNil(t, payload.Dispatch)
	assert.Equal(t, "queued", payload.Dispatch.Status)
	assert.Equal(t, "task-replay-123", payload.Dispatch.DispatchReference)
}

func TestRetryPostAdjudicationExecution_FailsWhenDeadLetterEvidenceMissing(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-replay-missing")

	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              receipts.EscrowAdjudicationRelease,
		Reason:               "release adjudicated",
	})
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, replayToolConfig(), store, nil, nil, nil, nil, nil, nil, &fakeAdjudicationBackgroundDispatcher{}), "retry_post_adjudication_execution")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "dead-letter")
}
