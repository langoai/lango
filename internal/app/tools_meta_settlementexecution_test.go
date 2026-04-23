package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/settlementexecution"
)

type fakeSettlementExecutionRuntime struct {
	err  error
	last settlementexecution.DirectPaymentRequest
}

func (f *fakeSettlementExecutionRuntime) ExecuteSettlement(_ context.Context, req settlementexecution.DirectPaymentRequest) (settlementexecution.DirectPaymentResult, error) {
	f.last = req
	if f.err != nil {
		return settlementexecution.DirectPaymentResult{}, f.err
	}
	return settlementexecution.DirectPaymentResult{Reference: "settlement-tx-123"}, nil
}

func TestBuildMetaTools_IncludesExecuteSettlement(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), nil, &fakeSettlementExecutionRuntime{}, nil, nil, nil, nil)
	tool := findTool(tools, "execute_settlement")
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

func TestBuildMetaTools_OmitsExecuteSettlementWithoutRuntime(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	require.Nil(t, findTool(tools, "execute_settlement"))
}

func TestExecuteSettlement_ApprovedPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-execute-settlement")
	runtime := &fakeSettlementExecutionRuntime{}

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, runtime, nil, nil, nil, nil), "execute_settlement")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(executeSettlementReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionSettled), payload.SettlementProgressionStatus)
	assert.Equal(t, "0.50", payload.ResolvedAmount)
	assert.Equal(t, "settlement-tx-123", payload.RuntimeReference)
	assert.Equal(t, settlementexecution.DirectPaymentRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		Counterparty:         "did:lango:peer-deal-execute-settlement",
		Amount:               "0.50",
	}, runtime.last)
}

func TestExecuteSettlement_RejectsWhenProgressionIsNotApprovedForSettlement(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-execute-settlement-review-needed")

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, &fakeSettlementExecutionRuntime{}, nil, nil, nil, nil), "execute_settlement")
	require.NotNil(t, tool)

	_, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
}

func TestExecuteSettlement_PropagatesRuntimeFailure(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-execute-settlement-runtime-failure")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, &fakeSettlementExecutionRuntime{err: errors.New("rpc timeout")}, nil, nil, nil, nil), "execute_settlement")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
}
