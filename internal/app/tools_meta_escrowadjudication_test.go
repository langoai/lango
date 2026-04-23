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

func TestBuildMetaTools_IncludesAdjudicateEscrowDispute(t *testing.T) {
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore()), "adjudicate_escrow_dispute")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := props["transaction_receipt_id"]
	_, hasOutcome := props["outcome"]
	assert.True(t, hasTransactionReceiptID)
	assert.True(t, hasOutcome)

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"transaction_receipt_id", "outcome"}, required)
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
