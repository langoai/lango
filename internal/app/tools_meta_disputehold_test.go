package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/disputehold"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/receipts"
)

type fakeDisputeHoldRuntime struct {
	err  error
	last disputehold.EscrowHoldRequest
}

func (f *fakeDisputeHoldRuntime) Hold(_ context.Context, req disputehold.EscrowHoldRequest) (disputehold.HoldResult, error) {
	f.last = req
	if f.err != nil {
		return disputehold.HoldResult{}, f.err
	}
	return disputehold.HoldResult{Reference: "hold-123"}, nil
}

func bindDisputeHoldEscrowExecutionInput(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
	t.Helper()

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "dispute hold test",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
}

func TestBuildMetaTools_IncludesHoldEscrowForDispute(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), nil, nil, nil, &fakeDisputeHoldRuntime{}, nil, nil)
	tool := findTool(tools, "hold_escrow_for_dispute")
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

func TestBuildMetaTools_OmitsHoldEscrowForDisputeWithoutRuntime(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		nil,
		receipts.NewStore(),
		escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig()),
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	require.Nil(t, findTool(tools, "hold_escrow_for_dispute"))
}

func TestHoldEscrowForDispute_DisputeReadyFundedPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold")
	runtime := &fakeDisputeHoldRuntime{}

	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, runtime, nil, nil), "hold_escrow_for_dispute")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(holdEscrowForDisputeReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionDisputeReady), payload.SettlementProgressionStatus)
	assert.Equal(t, "escrow-123", payload.EscrowReference)
	assert.Equal(t, "hold-123", payload.RuntimeReference)
	assert.Equal(t, disputehold.EscrowHoldRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
	}, runtime.last)
}

func TestHoldEscrowForDispute_RejectsWhenEscrowOrSettlementStateIsWrong(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cases := []struct {
		name       string
		setup      func(*testing.T, *receipts.Store, context.Context, receipts.TransactionReceipt)
		wantErrMsg string
	}{
		{
			name: "escrow not funded",
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
				bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
				require.NoError(t, err)
			},
			wantErrMsg: "escrow_not_funded",
		},
		{
			name: "settlement not dispute-ready",
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
				bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
				require.NoError(t, err)
			},
			wantErrMsg: "not_dispute_ready",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := receipts.NewStore()
			tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-"+tt.name)
			tt.setup(t, store, ctx, tx)

			tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, &fakeDisputeHoldRuntime{}, nil, nil), "hold_escrow_for_dispute")
			require.NotNil(t, tool)

			_, err := tool.Handler(ctx, map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrMsg)
		})
	}
}

func TestHoldEscrowForDispute_PropagatesRuntimeFailure(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-runtime-failure")

	bindDisputeHoldEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, &fakeDisputeHoldRuntime{err: errors.New("hold failed")}, nil, nil), "hold_escrow_for_dispute")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
}
