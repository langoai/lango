package app

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
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
	listProps, ok := listTool.Parameters["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, listProps, "manual_replay_actor")
	assert.Contains(t, listProps, "dead_lettered_after")
	assert.Contains(t, listProps, "dead_lettered_before")
	assert.Contains(t, listProps, "dead_letter_reason_query")
	assert.Contains(t, listProps, "latest_dispatch_reference")
	assert.Contains(t, listProps, "latest_status_subtype")
	assert.Contains(t, listProps, "manual_retry_count_min")
	assert.Contains(t, listProps, "manual_retry_count_max")
	assert.Contains(t, listProps, "total_retry_count_min")
	assert.Contains(t, listProps, "total_retry_count_max")
	assert.Contains(t, listProps, "latest_status_subtype_family")
	assert.Contains(t, listProps, "any_match_family")
	assert.Contains(t, listProps, "sort_by")

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
	assert.Equal(t, 1, payload["total"])
	assert.Equal(t, 0, payload["offset"])
	assert.Equal(t, 0, payload["limit"])

	entries := decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, deadTx.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.True(t, entries[0].IsDeadLettered)
	assert.True(t, entries[0].CanRetry)
	assert.Equal(t, string(receipts.EscrowAdjudicationRelease), entries[0].Adjudication)
}

func TestListDeadLetteredPostAdjudicationExecutions_AppliesFiltersAndPagination(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()

	releaseHigh := makeDeadLetterStatusFixture(t, store, ctxkeys.WithPrincipal(ctx, "operator:alice"), "status-release-high", receipts.EscrowAdjudicationRelease, 5, "dispatch-release-high")
	makeDeadLetterStatusFixture(t, store, ctx, "status-release-low", receipts.EscrowAdjudicationRelease, 2, "dispatch-release-low")
	refundHigh := makeDeadLetterStatusFixture(t, store, ctx, "status-refund-high", receipts.EscrowAdjudicationRefund, 4, "dispatch-refund-high")

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "list_dead_lettered_post_adjudication_executions")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"adjudication":             "release",
		"retry_attempt_min":        float64(4),
		"query":                    releaseHigh.TransactionReceiptID[:8],
		"dead_letter_reason_query": "release path",
		"offset":                   float64(0),
		"limit":                    float64(1),
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 1, payload["total"])
	assert.Equal(t, 0, payload["offset"])
	assert.Equal(t, 1, payload["limit"])

	entries := decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, releaseHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.Equal(t, 5, entries[0].LatestRetryAttempt)
	assert.Contains(t, entries[0].LatestDeadLetterReason, "release path")

	got, err = tool.Handler(ctx, map[string]interface{}{
		"adjudication":              "release",
		"retry_attempt_min":         float64(4),
		"query":                     releaseHigh.TransactionReceiptID[:8],
		"dead_letter_reason_query":  "release path",
		"latest_dispatch_reference": "dispatch-release-high",
	})
	require.NoError(t, err)

	payload, ok = got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 1, payload["total"])

	entries = decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, releaseHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.Equal(t, "dispatch-release-high", entries[0].LatestDispatchReference)
	assert.Equal(t, 1, entries[0].ManualRetryCount)
	assert.NotEmpty(t, entries[0].LatestManualReplayAt)
	assert.Equal(t, "dead-lettered", entries[0].LatestStatusSubtype)

	got, err = tool.Handler(ctx, map[string]interface{}{
		"latest_status_subtype":  "dead-lettered",
		"manual_retry_count_min": float64(1),
		"manual_retry_count_max": float64(1),
		"sort_by":                "latest_manual_replay_at",
	})
	require.NoError(t, err)

	payload, ok = got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 1, payload["total"])

	entries = decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, releaseHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.Equal(t, 1, entries[0].ManualRetryCount)
	assert.Equal(t, 3, entries[0].TotalRetryCount)
	assert.Equal(t, "dead-lettered", entries[0].LatestStatusSubtype)
	assert.Equal(t, "dead-letter", entries[0].LatestStatusSubtypeFamily)

	got, err = tool.Handler(ctx, map[string]interface{}{
		"adjudication":                 "release",
		"query":                        releaseHigh.TransactionReceiptID[:8],
		"latest_status_subtype_family": "dead-letter",
		"total_retry_count_min":        float64(3),
		"total_retry_count_max":        float64(3),
	})
	require.NoError(t, err)

	payload, ok = got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 1, payload["total"])

	entries = decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, releaseHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.Equal(t, 3, entries[0].TotalRetryCount)
	assert.Equal(t, "dead-letter", entries[0].LatestStatusSubtypeFamily)
	assert.ElementsMatch(t, []string{"retry", "manual-retry", "dead-letter"}, entries[0].AnyMatchFamilies)

	got, err = tool.Handler(ctx, map[string]interface{}{
		"any_match_family": "manual-retry",
	})
	require.NoError(t, err)

	payload, ok = got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 1, payload["total"])

	entries = decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, releaseHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
	assert.ElementsMatch(t, []string{"retry", "manual-retry", "dead-letter"}, entries[0].AnyMatchFamilies)

	got, err = tool.Handler(ctx, map[string]interface{}{
		"offset": float64(1),
		"limit":  float64(1),
	})
	require.NoError(t, err)

	payload, ok = got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, 3, payload["total"])
	assert.Equal(t, 1, payload["offset"])
	assert.Equal(t, 1, payload["limit"])

	entries = decodeDeadLetterEntriesFromPayload(t, payload["entries"])
	require.Len(t, entries, 1)
	assert.Equal(t, refundHigh.TransactionReceiptID, entries[0].TransactionReceiptID)
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
	assert.True(t, status.IsDeadLettered)
	assert.True(t, status.CanRetry)
	assert.Equal(t, string(receipts.EscrowAdjudicationRelease), status.Adjudication)
}

func decodeDeadLetterEntriesFromPayload(t *testing.T, value interface{}) []postadjudicationstatus.DeadLetterBacklogEntry {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	var entries []postadjudicationstatus.DeadLetterBacklogEntry
	require.NoError(t, json.Unmarshal(data, &entries))
	return entries
}

func makeDeadLetterStatusFixture(
	t *testing.T,
	store *receipts.Store,
	ctx context.Context,
	label string,
	outcome receipts.EscrowAdjudicationDecision,
	attempt int,
	dispatchReference string,
) receipts.TransactionReceipt {
	t.Helper()

	tx := createSubmittedTransaction(t, store, ctx, label)
	escrowReference := "escrow-" + label
	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, escrowReference, receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      escrowReference,
		RuntimeReference:     "hold-" + label,
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      escrowReference,
		Outcome:              outcome,
		Reason:               string(outcome) + " adjudicated",
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, receipts.PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              outcome,
		AttemptCount:         attempt - 1,
		NextRetryAt:          time.Now().UTC().Add(time.Minute),
		DispatchReference:    dispatchReference,
	})
	require.NoError(t, err)
	if actor := ctxkeys.PrincipalFromContext(ctx); actor != "" {
		err = store.RecordManualRetryRequested(ctx, receipts.ManualRetryRequestedRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			Outcome:              outcome,
			Reason:               "manual retry requested",
		})
		require.NoError(t, err)
	}
	err = store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              outcome,
		AttemptCount:         attempt,
		Reason:               "worker exhausted after " + strconv.Itoa(attempt) + " attempts on " + string(outcome) + " path",
	})
	require.NoError(t, err)
	return tx
}
