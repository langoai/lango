package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaTools_IncludesPostAdjudicationStatus(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())

	listTool := findTool(tools, "list_dead_lettered_post_adjudication_executions")
	require.NotNil(t, listTool)
	assert.Equal(t, "knowledge", listTool.Capability.Category)
	assert.Equal(t, agent.ActivityQuery, listTool.Capability.Activity)
	assert.True(t, listTool.Capability.ReadOnly)

	detailTool := findTool(tools, "get_post_adjudication_execution_status")
	require.NotNil(t, detailTool)
	assert.Equal(t, "knowledge", detailTool.Capability.Category)
	assert.Equal(t, agent.ActivityQuery, detailTool.Capability.Activity)
	assert.True(t, detailTool.Capability.ReadOnly)
}

func TestListDeadLetteredPostAdjudicationExecutions_ReturnsCurrentBacklogOnly(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()

	deadTx := createSubmittedTransaction(t, store, ctx, "deal-status-dead")
	bindDisputeHoldEscrowExecutionInput(t, store, ctx, deadTx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, deadTx.TransactionReceiptID, deadTx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, deadTx.TransactionReceiptID, deadTx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-dead-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, deadTx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, deadTx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: deadTx.TransactionReceiptID,
		SubmissionReceiptID:  deadTx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-dead-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: deadTx.TransactionReceiptID,
		SubmissionReceiptID:  deadTx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-dead-123",
		Outcome:              receipts.EscrowAdjudicationRelease,
		Reason:               "release adjudicated",
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: deadTx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRelease,
		AttemptCount:         4,
		Reason:               "worker exhausted",
	})
	require.NoError(t, err)

	recoveredTx := createSubmittedTransaction(t, store, ctx, "deal-status-recovered")
	bindDisputeHoldEscrowExecutionInput(t, store, ctx, recoveredTx)
	_, err = store.ApplyEscrowExecutionProgress(ctx, recoveredTx.TransactionReceiptID, recoveredTx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, recoveredTx.TransactionReceiptID, recoveredTx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-ok-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, recoveredTx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, recoveredTx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: recoveredTx.TransactionReceiptID,
		SubmissionReceiptID:  recoveredTx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-ok-123",
		RuntimeReference:     "hold-456",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: recoveredTx.TransactionReceiptID,
		SubmissionReceiptID:  recoveredTx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-ok-123",
		Outcome:              receipts.EscrowAdjudicationRefund,
		Reason:               "refund adjudicated",
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: recoveredTx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRefund,
		AttemptCount:         4,
		Reason:               "worker exhausted",
	})
	require.NoError(t, err)
	err = store.RecordManualRetryRequested(ctx, receipts.ManualRetryRequestedRequest{
		TransactionReceiptID: recoveredTx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRefund,
		Reason:               "manual retry requested",
	})
	require.NoError(t, err)

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "list_dead_lettered_post_adjudication_executions")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
}

func TestGetPostAdjudicationExecutionStatus_ReturnsCanonicalSnapshotAndSummary(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-status-detail")
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
	err = store.RecordPostAdjudicationRetryScheduled(ctx, receipts.PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRelease,
		AttemptCount:         2,
		NextRetryAt:          time.Now().UTC().Add(2 * time.Minute),
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRelease,
		AttemptCount:         3,
		Reason:               "terminal worker failure",
	})
	require.NoError(t, err)

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "get_post_adjudication_execution_status")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	status, ok := got.(postadjudicationstatus.TransactionStatus)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, status.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, 3, status.RetryDeadLetterSummary.LatestRetryAttempt)
	assert.Equal(t, "terminal worker failure", status.RetryDeadLetterSummary.LatestDeadLetterReason)
}
