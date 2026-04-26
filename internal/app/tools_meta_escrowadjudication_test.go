package app

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
)

type fakeAdjudicationBackgroundDispatcher struct {
	mu     sync.Mutex
	taskID string
	tasks  []background.TaskSnapshot
	prompt string
	origin background.Origin
	err    error
	calls  int
}

func (f *fakeAdjudicationBackgroundDispatcher) Submit(_ context.Context, prompt string, origin background.Origin) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.prompt = prompt
	f.origin = origin
	if f.err != nil {
		return "", f.err
	}
	if f.taskID == "" {
		f.taskID = "task-bg-dispatch-123"
	}
	return f.taskID, nil
}

func (f *fakeAdjudicationBackgroundDispatcher) snapshot() (calls int, prompt string, origin background.Origin) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls, f.prompt, f.origin
}

func (f *fakeAdjudicationBackgroundDispatcher) List() []background.TaskSnapshot {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]background.TaskSnapshot, len(f.tasks))
	copy(out, f.tasks)
	return out
}

func TestBuildMetaTools_IncludesAdjudicateEscrowDispute(t *testing.T) {
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore()), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)
	assert.Contains(t, tool.Description, "manual recovery")

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	_, hasOutcome := props["outcome"]
	_, hasAutoExecute := props["auto_execute"]
	_, hasBackgroundExecute := props["background_execute"]
	assert.True(t, hasTransactionReceiptID)
	assert.True(t, hasOutcome)
	assert.True(t, hasAutoExecute)
	assert.True(t, hasBackgroundExecute)

	autoExecute, ok := props["auto_execute"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, autoExecute["description"], "manual recovery")

	backgroundExecute, ok := props["background_execute"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, backgroundExecute["description"], "manual recovery")

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"transaction_receipt_id", "outcome"}, required)
}

func TestBuildMetaTools_AdjudicateEscrowDisputeDocumentsManualRecoveryDefault(t *testing.T) {
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore()), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	assert.Contains(t, tool.Description, "manual recovery")

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	autoExecute, _ := props["auto_execute"].(map[string]interface{})
	backgroundExecute, _ := props["background_execute"].(map[string]interface{})

	autoDescription, _ := autoExecute["description"].(string)
	backgroundDescription, _ := backgroundExecute["description"].(string)

	assert.Contains(t, autoDescription, "inline")
	assert.Contains(t, backgroundDescription, "background")
	assert.Contains(t, backgroundDescription, "manual recovery")
}

func TestAdjudicateEscrowDispute_BackgroundExecuteReturnsDispatchReceipt(t *testing.T) {
	t.Parallel()

	ctx := session.WithSessionKey(context.Background(), "telegram:chat-123:user-456")
	cases := []struct {
		name    string
		outcome string
	}{
		{
			name:    "release",
			outcome: "release",
		},
		{
			name:    "refund",
			outcome: "refund",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := receipts.NewStore()
			tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-background-"+tt.name)

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

			dispatcher := &fakeAdjudicationBackgroundDispatcher{taskID: "task-bg-dispatch-123"}
			tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, nil, dispatcher), "adjudicate_escrow_dispute")
			require.NotNil(t, tool)

			got, err := tool.Handler(ctx, map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
				"outcome":                tt.outcome,
				"background_execute":     true,
			})
			require.NoError(t, err)

			payload, ok := got.(adjudicateEscrowDisputeReceipt)
			require.True(t, ok)
			assert.Equal(t, tt.outcome, payload.Outcome)
			require.NotNil(t, payload.BackgroundDispatchReceipt)
			assert.Nil(t, payload.Execution)
			assert.Equal(t, "queued", payload.BackgroundDispatchReceipt.Status)
			assert.Equal(t, tx.TransactionReceiptID, payload.BackgroundDispatchReceipt.TransactionReceiptID)
			assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.BackgroundDispatchReceipt.SubmissionReceiptID)
			assert.Equal(t, "escrow-123", payload.BackgroundDispatchReceipt.EscrowReference)
			assert.Equal(t, tt.outcome, payload.BackgroundDispatchReceipt.Outcome)
			assert.Equal(t, "task-bg-dispatch-123", payload.BackgroundDispatchReceipt.DispatchReference)
			calls, prompt, origin := dispatcher.snapshot()
			assert.Equal(t, 1, calls)
			assert.Contains(t, prompt, "transaction_receipt_id="+tx.TransactionReceiptID)
			if tt.outcome == "release" {
				assert.Contains(t, prompt, "release_escrow_settlement")
			} else {
				assert.Contains(t, prompt, "refund_escrow_settlement")
			}
			assert.Equal(t, "telegram:chat-123", origin.Channel)
			assert.Equal(t, "telegram:chat-123:user-456", origin.Session)
		})
	}
}

func TestAdjudicateEscrowDispute_RejectsMutuallyExclusiveExecutionModes(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	tx := createSubmittedTransaction(t, store, context.Background(), "deal-escrow-adjudication-mode-conflict")

	bindDisputeHoldEscrowExecutionInput(t, store, context.Background(), tx)
	_, err := store.ApplyEscrowExecutionProgress(context.Background(), tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(context.Background(), tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(context.Background(), tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(context.Background(), tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(context.Background(), receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	_, err = tool.Handler(context.Background(), map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
		"auto_execute":           true,
		"background_execute":     true,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "mutually exclusive")
}

func TestAdjudicateEscrowDispute_DefaultsToManualRecoveryWhenExecutionFlagsAreAbsent(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-manual-default")

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

	dispatcher := &fakeAdjudicationBackgroundDispatcher{taskID: "task-bg-dispatch-123"}
	releaseRuntime := &fakeEscrowReleaseRuntime{}
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, releaseRuntime, nil, dispatcher), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
	})
	require.NoError(t, err)

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, "release", payload.Outcome)
	assert.Nil(t, payload.Execution)
	assert.Nil(t, payload.BackgroundDispatchReceipt)

	calls, _, _ := dispatcher.snapshot()
	assert.Equal(t, 0, calls)
	assert.Equal(t, escrowrelease.ReleaseRequest{}, releaseRuntime.last)
}

func TestAdjudicateEscrowDispute_DisputeHeldFundedPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication")

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

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
		"reason":                 "fulfilled after review",
	})
	require.NoError(t, err)

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionApprovedForSettlement), payload.SettlementProgressionStatus)
	assert.Equal(t, "escrow-123", payload.EscrowReference)
	assert.Equal(t, "release", payload.Outcome)
	assert.Nil(t, payload.Execution)
	assert.Nil(t, payload.BackgroundDispatchReceipt)
}

func TestAdjudicateEscrowDispute_RejectsWrongStateOrOutcome(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cases := []struct {
		name       string
		params     map[string]interface{}
		setup      func(*testing.T, *receipts.Store, context.Context, receipts.TransactionReceipt)
		wantErrMsg string
	}{
		{
			name: "hold missing",
			params: map[string]interface{}{
				"outcome": "release",
			},
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
				bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
				require.NoError(t, err)
			},
			wantErrMsg: "hold_evidence_missing",
		},
		{
			name: "invalid outcome",
			params: map[string]interface{}{
				"outcome": "other",
			},
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
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
			},
			wantErrMsg: "invalid_outcome",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := receipts.NewStore()
			tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-"+tt.name)
			tt.setup(t, store, ctx, tx)

			params := map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
			}
			for k, v := range tt.params {
				params[k] = v
			}

			tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "adjudicate_escrow_dispute")
			require.NotNil(t, tool)
			_, err := tool.Handler(ctx, params)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrMsg)
		})
	}
}

func TestAdjudicateEscrowDispute_AutoExecuteReleaseReturnsNestedExecutionResult(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-auto-release")

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

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, &fakeEscrowReleaseRuntime{}, nil), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
		"auto_execute":           true,
	})
	require.NoError(t, err)

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	require.NotNil(t, payload.Execution)
	assert.Equal(t, "release", payload.Execution.Branch)
	assert.Equal(t, string(escrowrelease.StatusSettledTarget), payload.Execution.Status)
	assert.Equal(t, string(receipts.SettlementProgressionSettled), payload.Execution.SettlementProgressionStatus)
	assert.Equal(t, "0.50", payload.Execution.ResolvedAmount)
	assert.Equal(t, "escrow-release-tx-123", payload.Execution.RuntimeReference)
}

func TestAdjudicateEscrowDispute_AutoExecuteRefundReturnsNestedExecutionResult(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-auto-refund")

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

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, &fakeEscrowRefundRuntime{}), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "refund",
		"auto_execute":           true,
	})
	require.NoError(t, err)

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	require.NotNil(t, payload.Execution)
	assert.Equal(t, "refund", payload.Execution.Branch)
	assert.Equal(t, string(escrowrefund.StatusRefundExecuted), payload.Execution.Status)
	assert.Equal(t, string(receipts.SettlementProgressionReviewNeeded), payload.Execution.SettlementProgressionStatus)
	assert.Equal(t, "0.50", payload.Execution.ResolvedAmount)
	assert.Equal(t, "refund-tx-123", payload.Execution.RuntimeReference)
}

func TestAdjudicateEscrowDispute_AutoExecuteFailureStillReturnsAdjudicationReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-auto-release-failure")

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

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, &fakeEscrowReleaseRuntime{err: errors.New("release failed")}, nil), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
		"auto_execute":           true,
	})
	require.Error(t, err)

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, "release", payload.Outcome)
	require.NotNil(t, payload.Execution)
	assert.Equal(t, "release", payload.Execution.Branch)
	assert.Equal(t, string(escrowrelease.StatusFailed), payload.Execution.Status)

	updated, getErr := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, getErr)
	assert.Equal(t, receipts.EscrowAdjudicationRelease, updated.EscrowAdjudication)
	assert.Equal(t, receipts.SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
}
