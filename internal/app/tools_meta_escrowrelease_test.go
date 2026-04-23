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
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/receipts"
)

type fakeEscrowReleaseRuntime struct {
	err  error
	last escrowrelease.ReleaseRequest
}

func (f *fakeEscrowReleaseRuntime) Release(_ context.Context, req escrowrelease.ReleaseRequest) (escrowrelease.ReleaseResult, error) {
	f.last = req
	if f.err != nil {
		return escrowrelease.ReleaseResult{}, f.err
	}
	return escrowrelease.ReleaseResult{Reference: "escrow-release-tx-123"}, nil
}

func bindReleaseEscrowExecutionInput(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
	t.Helper()

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow release test",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
}

func TestBuildMetaTools_IncludesReleaseEscrowSettlement(t *testing.T) {
	tools := buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), nil, nil, nil, &fakeEscrowReleaseRuntime{}, nil)
	tool := findTool(tools, "release_escrow_settlement")
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

func TestBuildMetaTools_OmitsReleaseEscrowSettlementWithoutRuntime(t *testing.T) {
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
	)
	require.Nil(t, findTool(tools, "release_escrow_settlement"))
}

func TestReleaseEscrowSettlement_FundedApprovedPathReturnsCanonicalReceipt(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-release-escrow")
	runtime := &fakeEscrowReleaseRuntime{}

	bindReleaseEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, runtime, nil), "release_escrow_settlement")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(releaseEscrowSettlementReceipt)
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, tx.CurrentSubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, string(receipts.SettlementProgressionSettled), payload.SettlementProgressionStatus)
	assert.Equal(t, "0.50", payload.ResolvedAmount)
	assert.Equal(t, "escrow-release-tx-123", payload.RuntimeReference)
	assert.Equal(t, escrowrelease.ReleaseRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Amount:               "0.50",
	}, runtime.last)
}

func TestReleaseEscrowSettlement_RejectsWhenEscrowOrSettlementStateIsWrong(t *testing.T) {
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
				bindReleaseEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
				require.NoError(t, err)
			},
			wantErrMsg: "escrow_not_funded",
		},
		{
			name: "settlement not approved",
			setup: func(t *testing.T, store *receipts.Store, ctx context.Context, tx receipts.TransactionReceipt) {
				t.Helper()
				bindReleaseEscrowExecutionInput(t, store, ctx, tx)
				_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
				require.NoError(t, err)
				_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
				require.NoError(t, err)
				_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionReviewNeeded, receipts.SettlementProgressionReasonCodeRequestRevision, "review needed", "")
				require.NoError(t, err)
			},
			wantErrMsg: "not_approved_for_settlement",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := receipts.NewStore()
			tx := createSubmittedTransaction(t, store, ctx, "deal-release-escrow-"+tt.name)
			tt.setup(t, store, ctx, tx)

			tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, &fakeEscrowReleaseRuntime{}, nil), "release_escrow_settlement")
			require.NotNil(t, tool)

			_, err := tool.Handler(ctx, map[string]interface{}{
				"transaction_receipt_id": tx.TransactionReceiptID,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrMsg)
		})
	}
}

func TestReleaseEscrowSettlement_PropagatesRuntimeFailure(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	ctx := context.Background()
	tx := createSubmittedTransaction(t, store, ctx, "deal-release-escrow-runtime-failure")

	bindReleaseEscrowExecutionInput(t, store, ctx, tx)
	_, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, "", receipts.EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, "escrow-123", receipts.EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, receipts.SettlementProgressionApprovedForSettlement, receipts.SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	tool := findTool(buildMetaToolsWithRuntimes(nil, nil, nil, config.SkillConfig{}, nil, store, nil, nil, nil, &fakeEscrowReleaseRuntime{err: errors.New("release failed")}, nil), "release_escrow_settlement")
	require.NotNil(t, tool)

	_, err = tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.Error(t, err)
}
