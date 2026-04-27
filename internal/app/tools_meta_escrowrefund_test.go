package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/receipts"
)

type fakeEscrowRefundRuntime struct {
	err  error
	last escrowrefund.RefundRequest
}

func (f *fakeEscrowRefundRuntime) Refund(_ context.Context, req escrowrefund.RefundRequest) (escrowrefund.RefundResult, error) {
	f.last = req
	if f.err != nil {
		return escrowrefund.RefundResult{}, f.err
	}
	return escrowrefund.RefundResult{Reference: "refund-tx-123"}, nil
}

func TestBuildMetaTools_IncludesRefundEscrowSettlement(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), nil, nil, nil, nil, nil, &fakeEscrowRefundRuntime{})
	tool := findTool(tools, "refund_escrow_settlement")
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

func TestBuildMetaTools_OmitsRefundEscrowSettlementWithoutRuntime(t *testing.T) {
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
	require.Nil(t, findTool(tools, "refund_escrow_settlement"))
}

func bindRefundEscrowExecutionInput(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
	t.Helper()

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow refund test",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
}

func adjudicateRefundEscrow(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
	t.Helper()

	err := store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
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
		Outcome:              receipts.EscrowAdjudicationRefund,
		Reason:               "refund adjudicated",
	})
	require.NoError(t, err)
}

func TestRefundEscrowSettlement_FundedReviewNeededPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-refund-escrow")
	runtime := &fakeEscrowRefundRuntime{}

	bindRefundEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	adjudicateRefundEscrow(t, store, ctx, tx)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, runtime), "refund_escrow_settlement")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(refundEscrowSettlementReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionReviewNeeded), payload.SettlementProgressionStatus)
	assert.Equal(t, "0.50", payload.ResolvedAmount)
	assert.Equal(t, "refund-tx-123", payload.RuntimeReference)
	assert.Equal(t, escrowrefund.RefundRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Amount:               "0.50",
	}, runtime.last)
}

func TestRefundEscrowSettlement_RejectsWhenEscrowOrSettlementStateIsWrong(t *testing.T) {
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
				bindRefundEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "refund review", "")
				require.NoError(t, err)
			},
			wantErrMsg: "escrow_not_funded",
		},
		{
			name: "settlement not review-needed",
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
				bindRefundEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
				require.NoError(t, err)
			},
			wantErrMsg: "not_review_needed",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := receipts.NewStore()
			tx := createSubmittedTransaction(t, store, ctx, "deal-refund-escrow-"+tt.name)
			tt.setup(t, store, ctx, tx)

			tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, &fakeEscrowRefundRuntime{}), "refund_escrow_settlement")
			require.NotNil(t, tool)

			_, err := tool.Handler(ctx, map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrMsg)
		})
	}
}

func TestRefundEscrowSettlement_PropagatesRuntimeFailure(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-refund-escrow-runtime-failure")

	bindRefundEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionDisputeReady, receipts.SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	adjudicateRefundEscrow(t, store, ctx, tx)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, nil, nil, &fakeEscrowRefundRuntime{err: errors.New("refund failed")}), "refund_escrow_settlement")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
}
