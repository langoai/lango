package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	engine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, config.DefaultConfig(), store, engine)

	for _, name := range []string{
		"execute_escrow_recommendation",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
	} {
		assert.NotNil(t, findTool(tools, name), "expected %s to be registered", name)
	}
}

func TestRuntimeBackedMetaToolsRejectWhitespaceTransactionReceiptID(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		config.DefaultConfig(),
		receipts.NewStore(),
		fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime{},
		&fakeSettlementExecutionRuntime{},
		&fakePartialSettlementExecutionRuntime{},
		&fakeDisputeHoldRuntime{},
		&fakeEscrowReleaseRuntime{},
		&fakeEscrowRefundRuntime{},
		&fakeAdjudicationBackgroundDispatcher{},
	)

	for _, name := range []string{
		"execute_settlement",
		"execute_partial_settlement",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
		"get_post_adjudication_execution_status",
		"retry_post_adjudication_execution",
	} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tool := findTool(tools, name)
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), map[string]interface{}{
				"transaction_receipt_id": " \t\n ",
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, "missing transaction_receipt_id parameter")
		})
	}
}

func TestAdjudicateEscrowDisputeBackgroundMissingDispatcherReturnsCanonicalReceiptAndError(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpersPrepareHeldEscrowTransaction(t, store, ctx, "deal-buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-background-missing")
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, nil), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "release",
		"background_execute":     true,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "background manager is not configured")

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, "escrow-123", payload.EscrowReference)
	assert.Equal(t, "release", payload.Outcome)
	assert.Nil(t, payload.BackgroundDispatchReceipt)
	assert.Nil(t, payload.Execution)
}

func TestAdjudicateEscrowDisputeBackgroundSubmitFailureReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpersPrepareHeldEscrowTransaction(t, store, ctx, "deal-buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-background-submit-failure")
	dispatcher := &fakeAdjudicationBackgroundDispatcher{err: errors.New("queue full")}
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, nil, dispatcher), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "refund",
		"background_execute":     true,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "submit background execution: queue full")

	payload, ok := got.(adjudicateEscrowDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, "refund", payload.Outcome)
	assert.Nil(t, payload.BackgroundDispatchReceipt)
	assert.Nil(t, payload.Execution)
	calls, prompt, _ := dispatcher.snapshot()
	assert.Equal(t, 1, calls)
	assert.Contains(t, prompt, "refund_escrow_settlement")
}

func TestAdjudicateEscrowDisputeInlineMissingRuntimeReturnsErrorAfterAdjudication(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpersPrepareHeldEscrowTransaction(t, store, ctx, "deal-buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-inline-missing-runtime")
	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, nil), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "refund",
		"auto_execute":           true,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "escrow refund runtime is not configured")

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.EscrowAdjudicationRefund, updated.EscrowAdjudication)
	assert.Equal(t, receipts.SettlementProgressionReviewNeeded, updated.SettlementProgressionStatus)
}

func TestEngineEscrowReleaseAndRefundRuntimeSurfaceEngineErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	engine := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())

	release, err := (engineEscrowReleaseRuntime{engine: engine}).Release(ctx, escrowrelease.ReleaseRequest{
		EscrowReference: "missing-release",
	})
	require.Error(t, err)
	assert.Equal(t, escrowrelease.ReleaseResult{}, release)
	assert.ErrorContains(t, err, "escrow not found")

	refund, err := (engineEscrowRefundRuntime{engine: engine}).Refund(ctx, escrowrefund.RefundRequest{
		EscrowReference: "missing-refund",
	})
	require.Error(t, err)
	assert.Equal(t, escrowrefund.RefundResult{}, refund)
	assert.ErrorContains(t, err, "escrow not found")
}

func buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpersPrepareHeldEscrowTransaction(t *testing.T, store *receipts.Store, ctx context.Context, transactionID string) receipts.TransactionReceipt {
	t.Helper()

	tx := createSubmittedTransaction(t, store, ctx, transactionID)
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
	return tx
}
