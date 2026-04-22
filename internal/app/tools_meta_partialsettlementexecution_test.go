package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/partialsettlementexecution"
	"github.com/langoai/lango/internal/receipts"
)

type fakePartialSettlementExecutionRuntime struct {
	err  error
	last partialsettlementexecution.DirectPaymentRequest
}

func (f *fakePartialSettlementExecutionRuntime) ExecuteSettlement(_ context.Context, req partialsettlementexecution.DirectPaymentRequest) (partialsettlementexecution.DirectPaymentResult, error) {
	f.last = req
	if f.err != nil {
		return partialsettlementexecution.DirectPaymentResult{}, f.err
	}
	return partialsettlementexecution.DirectPaymentResult{Reference: "partial-settlement-tx-123"}, nil
}

func TestBuildMetaTools_IncludesExecutePartialSettlement(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), nil, nil, &fakePartialSettlementExecutionRuntime{}, nil)
	tool := findTool(tools, "execute_partial_settlement")
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

func TestBuildMetaTools_OmitsExecutePartialSettlementWithoutRuntime(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	require.Nil(t, findTool(tools, "execute_partial_settlement"))
}

func TestExecutePartialSettlement_ApprovedPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-execute-partial-settlement")
	runtime := &fakePartialSettlementExecutionRuntime{}

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "settle:0.40-usdc")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, runtime, nil), "execute_partial_settlement")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(executePartialSettlementReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionPartiallySettled), payload.SettlementProgressionStatus)
	assert.Equal(t, "0.40", payload.ExecutedAmount)
	assert.Equal(t, "0.10", payload.RemainingAmount)
	assert.Equal(t, "partial-settlement-tx-123", payload.RuntimeReference)
	assert.Equal(t, partialsettlementexecution.DirectPaymentRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		Counterparty:         "did:lango:peer-deal-execute-partial-settlement",
		Amount:               "0.40",
	}, runtime.last)
}

func TestExecutePartialSettlement_RejectsMissingOrInvalidPartialHint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		partialHint string
		wantErr     string
	}{
		{
			name:    "missing partial hint",
			wantErr: "partial_hint_missing",
		},
		{
			name:        "invalid partial hint",
			partialHint: "settle:forty%",
			wantErr:     "partial_hint_invalid",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := receipts.NewStore()
			ctx := context.Background()
			tx := createSubmittedTransaction(t, store, ctx, "deal-execute-partial-settlement-"+tt.name)
			_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", tt.partialHint)
			require.NoError(t, err)

			tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, &fakePartialSettlementExecutionRuntime{}, nil), "execute_partial_settlement")
			require.NotNil(t, tool)

			_, err = tool.Handler(ctx, map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
