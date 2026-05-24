package app

import (
	"context"
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

	calls, prompt, origin := dispatcher.snapshot()
	assert.Equal(t, 1, calls)
	assert.Contains(t, prompt, "release_escrow_settlement")
	assert.Contains(t, prompt, "transaction_receipt_id="+tx.TransactionReceiptID)
	assert.Contains(t, prompt, "submission_receipt_id="+tx.CurrentSubmissionReceiptID)
	assert.Contains(t, prompt, "escrow_reference=escrow-123")
	assert.Contains(t, prompt, "Do not re-adjudicate.")
	assert.Empty(t, origin.Channel)
	assert.Empty(t, origin.Session)
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

func TestRetryPostAdjudicationExecution_FailsWhenOnlyManualRetryEvidenceExists(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-replay-manual-only")

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
	err = store.RecordManualRetryRequested(ctx, receipts.ManualRetryRequestedRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              receipts.EscrowAdjudicationRelease,
		Reason:               "operator requested replay",
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

func TestRetryPostAdjudicationExecution_RequiresTransactionReceiptIDParameter(t *testing.T) {
	t.Parallel()

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, replayToolConfig(), receipts.NewStore(), nil, nil, nil, nil, nil, nil, &fakeAdjudicationBackgroundDispatcher{}), "retry_post_adjudication_execution")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing transaction_receipt_id parameter")
}

func TestListDeadLetteredPostAdjudicationExecutions_PassesAllFilterOptions(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	matchingTx := createSubmittedTransaction(t, store, ctx, "deal-list-dead-letter-matching")
	otherTx := createSubmittedTransaction(t, store, ctx, "deal-list-dead-letter-other")

	recordPostAdjudicationDeadLetterForListTest(t, store, ctx, matchingTx, receipts.EscrowAdjudicationRelease, "escrow-list-release", 5, "release worker failed", "operator:alice")
	recordPostAdjudicationDeadLetterForListTest(t, store, ctx, otherTx, receipts.EscrowAdjudicationRefund, "escrow-list-refund", 2, "refund worker failed", "")

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, nil), "list_dead_lettered_post_adjudication_executions")
	require.NotNil(t, tool)

	cases := []struct {
		name       string
		params     map[string]interface{}
		wantTotal  int
		wantCount  int
		wantTx     string
		wantOffset int
		wantLimit  int
	}{
		{
			name:      "adjudication",
			params:    map[string]interface{}{"adjudication": string(receipts.EscrowAdjudicationRelease)},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:      "retry attempt range",
			params:    map[string]interface{}{"retry_attempt_min": float64(5), "retry_attempt_max": float64(5)},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:      "query",
			params:    map[string]interface{}{"query": matchingTx.TransactionReceiptID},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:      "manual replay actor",
			params:    map[string]interface{}{"manual_replay_actor": "operator:alice"},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name: "dead letter time window",
			params: map[string]interface{}{
				"dead_lettered_after":  time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
				"dead_lettered_before": time.Now().Add(time.Minute).UTC().Format(time.RFC3339),
			},
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name:      "dead letter reason",
			params:    map[string]interface{}{"dead_letter_reason_query": "release WORKER"},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:      "latest subtype",
			params:    map[string]interface{}{"latest_status_subtype": "dead-lettered", "latest_status_subtype_family": "dead-letter"},
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name:      "retry totals",
			params:    map[string]interface{}{"manual_retry_count_min": float64(1), "manual_retry_count_max": float64(1), "total_retry_count_min": float64(2), "total_retry_count_max": float64(2)},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name: "transaction global retry family",
			params: map[string]interface{}{
				"transaction_global_total_retry_count_min": float64(2),
				"transaction_global_total_retry_count_max": float64(2),
				"transaction_global_any_match_family":      "manual-retry",
			},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:      "current submission family",
			params:    map[string]interface{}{"any_match_family": "manual-retry"},
			wantTotal: 1,
			wantCount: 1,
			wantTx:    matchingTx.TransactionReceiptID,
		},
		{
			name:       "pagination and sort",
			params:     map[string]interface{}{"sort_by": "latest_manual_replay_at", "offset": float64(1), "limit": float64(1)},
			wantTotal:  2,
			wantCount:  1,
			wantOffset: 1,
			wantLimit:  1,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(ctx, tt.params)
			require.NoError(t, err)

			payload, ok := got.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.wantCount, payload["count"])
			assert.Equal(t, tt.wantTotal, payload["total"])
			assert.Equal(t, tt.wantOffset, payload["offset"])
			assert.Equal(t, tt.wantLimit, payload["limit"])
			entries, ok := payload["entries"].([]postadjudicationstatus.DeadLetterBacklogEntry)
			require.True(t, ok)
			require.Len(t, entries, tt.wantCount)
			if tt.wantTx != "" {
				assert.Equal(t, tt.wantTx, entries[0].TransactionReceiptID)
				assert.Equal(t, "operator:alice", entries[0].LatestManualReplayActor)
				assert.Equal(t, 1, entries[0].ManualRetryCount)
				assert.Equal(t, 2, entries[0].TotalRetryCount)
				assert.Equal(t, "release worker failed", entries[0].LatestDeadLetterReason)
			}
		})
	}
}

func recordPostAdjudicationDeadLetterForListTest(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt, outcome receipts.EscrowAdjudicationDecision, escrowReference string, attempt int, reason, manualReplayActor string) {
	t.Helper()

	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, escrowReference, receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	require.NoError(t, store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      escrowReference,
		RuntimeReference:     "hold-" + escrowReference,
	}))
	_, err = store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      escrowReference,
		Outcome:              outcome,
		Reason:               string(outcome) + " adjudicated",
	})
	require.NoError(t, err)
	if manualReplayActor != "" {
		require.NoError(t, store.RecordManualRetryRequested(ctxkeys.WithPrincipal(ctx, manualReplayActor), receipts.ManualRetryRequestedRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			Outcome:              outcome,
			Reason:               "operator retry",
		}))
	}
	require.NoError(t, store.RecordPostAdjudicationDeadLetter(ctx, receipts.PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              outcome,
		AttemptCount:         attempt,
		Reason:               reason,
	}))
}
