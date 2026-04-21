package escrowexecution

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestService_ExecuteRecommendation_CreateAndFundSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx := createApprovedEscrowReceipt(t, ctx, store)
	runtime := &fakeRuntime{
		createEntry: &escrow.EscrowEntry{ID: "escrow-1"},
		fundEntry:   &escrow.EscrowEntry{ID: "escrow-1"},
	}

	service := NewService(store, runtime)
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	assert.Equal(t, tx.TransactionReceiptID, result.TransactionReceiptID)
	assert.Equal(t, submission.SubmissionReceiptID, result.SubmissionReceiptID)
	assert.Equal(t, "escrow-1", result.EscrowReference)
	assert.Equal(t, receipts.EscrowExecutionStatusFunded, result.EscrowExecutionStatus)

	require.Len(t, runtime.createCalls, 1)
	assert.Equal(t, "did:lango:buyer", runtime.createCalls[0].BuyerDID)
	assert.Equal(t, "did:lango:seller", runtime.createCalls[0].SellerDID)
	assert.Zero(t, runtime.createCalls[0].Amount.Cmp(mustBigInt(t, "25000000")))
	require.Len(t, runtime.createCalls[0].Milestones, 2)
	assert.Zero(t, runtime.createCalls[0].Milestones[0].Amount.Cmp(mustBigInt(t, "10000000")))
	assert.Zero(t, runtime.createCalls[0].Milestones[1].Amount.Cmp(mustBigInt(t, "15000000")))
	assert.Equal(t, []string{"escrow-1"}, runtime.fundCalls)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.EscrowExecutionStatusFunded, updatedTx.EscrowExecutionStatus)
	assert.Equal(t, "escrow-1", updatedTx.EscrowReference)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 4)
	assert.Equal(t, receipts.EventPaymentApproval, events[0].Type)
	assert.Equal(t, receipts.EventEscrowExecutionStarted, events[1].Type)
	assert.Equal(t, receipts.EventEscrowExecutionCreated, events[2].Type)
	assert.Equal(t, receipts.EventEscrowExecutionFunded, events[3].Type)
}

func TestService_ExecuteRecommendation_DeniesWhenApprovalIsNotApproved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-not-approved",
		ArtifactLabel:       "artifact/not-approved",
		PayloadHash:         "hash-not-approved",
		SourceLineageDigest: "lineage-not-approved",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionReject,
		Reason:        "Rejected.",
		SuggestedMode: paymentapproval.ModeReject,
	})
	require.NoError(t, err)
	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "25.00",
		Reason:    "knowledge exchange",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "delivery", Amount: "25.00"},
		},
	})
	require.NoError(t, err)

	service := NewService(store, &fakeRuntime{})
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment approval status")
	assert.Equal(t, Result{}, result)
}

func TestService_ExecuteRecommendation_DeniesWhenSettlementHintIsNotEscrow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-mode-mismatch",
		ArtifactLabel:       "artifact/mode-mismatch",
		PayloadHash:         "hash-mode-mismatch",
		SourceLineageDigest: "lineage-mode-mismatch",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	service := NewService(store, &fakeRuntime{})
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "settlement hint")
	assert.Equal(t, Result{}, result)
}

func TestService_ExecuteRecommendation_DeniesWhenInputIsNotBound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-input-missing",
		ArtifactLabel:       "artifact/input-missing",
		PayloadHash:         "hash-input-missing",
		SourceLineageDigest: "lineage-input-missing",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)

	service := NewService(store, &fakeRuntime{})
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "escrow execution input")
	assert.Equal(t, Result{}, result)
}

func TestService_ExecuteRecommendation_CreateFailureRecordsFailedState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx := createApprovedEscrowReceipt(t, ctx, store)
	runtime := &fakeRuntime{
		createEntry: &escrow.EscrowEntry{ID: "escrow-create-failed"},
		createErr:   errors.New("create escrow failed"),
	}

	service := NewService(store, runtime)
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create escrow")
	assert.Equal(t, Result{}, result)

	updatedTx, getErr := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, getErr)
	assert.Equal(t, receipts.EscrowExecutionStatusFailed, updatedTx.EscrowExecutionStatus)
	assert.Empty(t, updatedTx.EscrowReference)

	_, events, getErr := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, getErr)
	require.Len(t, events, 3)
	assert.Equal(t, receipts.EventEscrowExecutionStarted, events[1].Type)
	assert.Equal(t, receipts.EventEscrowExecutionFailed, events[2].Type)
	assert.Contains(t, events[2].Reason, "create escrow failed")
}

func TestService_ExecuteRecommendation_FundFailurePreservesCreatedReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	submission, tx := createApprovedEscrowReceipt(t, ctx, store)
	runtime := &fakeRuntime{
		createEntry: &escrow.EscrowEntry{ID: "escrow-created"},
		fundErr:     errors.New("lock funds failed"),
	}

	service := NewService(store, runtime)
	result, err := service.ExecuteRecommendation(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fund escrow")
	assert.Equal(t, Result{}, result)

	updatedTx, getErr := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, getErr)
	assert.Equal(t, receipts.EscrowExecutionStatusFailed, updatedTx.EscrowExecutionStatus)
	assert.Equal(t, "escrow-created", updatedTx.EscrowReference)

	_, events, getErr := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, getErr)
	require.Len(t, events, 4)
	assert.Equal(t, receipts.EventEscrowExecutionFailed, events[3].Type)
	assert.Contains(t, events[3].Reason, "lock funds failed")
}

func TestService_ExecuteRecommendation_RerunAfterFundedIsRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	_, tx := createApprovedEscrowReceipt(t, ctx, store)
	runtime := &fakeRuntime{
		createEntry: &escrow.EscrowEntry{ID: "escrow-1"},
		fundEntry:   &escrow.EscrowEntry{ID: "escrow-1"},
	}

	service := NewService(store, runtime)
	_, err := service.ExecuteRecommendation(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.NoError(t, err)

	result, err := service.ExecuteRecommendation(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already progressed")
	assert.Equal(t, Result{}, result)
	require.Len(t, runtime.createCalls, 1)
	require.Len(t, runtime.fundCalls, 1)
}

func TestService_ExecuteRecommendation_RerunAfterFailedIsRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := receipts.NewStore()
	_, tx := createApprovedEscrowReceipt(t, ctx, store)
	runtime := &fakeRuntime{
		createEntry: &escrow.EscrowEntry{ID: "escrow-created"},
		fundErr:     errors.New("lock funds failed"),
	}

	service := NewService(store, runtime)
	_, err := service.ExecuteRecommendation(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.Error(t, err)

	result, err := service.ExecuteRecommendation(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already progressed")
	assert.Equal(t, Result{}, result)
	require.Len(t, runtime.createCalls, 1)
	require.Len(t, runtime.fundCalls, 1)
}

type fakeRuntime struct {
	createEntry *escrow.EscrowEntry
	createErr   error
	fundEntry   *escrow.EscrowEntry
	fundErr     error
	createCalls []escrow.CreateRequest
	fundCalls   []string
}

func (f *fakeRuntime) Create(_ context.Context, req escrow.CreateRequest) (*escrow.EscrowEntry, error) {
	f.createCalls = append(f.createCalls, cloneCreateRequest(req))
	if f.createErr != nil {
		return f.createEntry, f.createErr
	}
	if f.createEntry == nil {
		return &escrow.EscrowEntry{ID: "escrow-default"}, nil
	}
	return f.createEntry, nil
}

func (f *fakeRuntime) Fund(_ context.Context, escrowID string) (*escrow.EscrowEntry, error) {
	f.fundCalls = append(f.fundCalls, escrowID)
	if f.fundErr != nil {
		return f.fundEntry, f.fundErr
	}
	if f.fundEntry == nil {
		return &escrow.EscrowEntry{ID: escrowID}, nil
	}
	return f.fundEntry, nil
}

func createApprovedEscrowReceipt(t *testing.T, ctx context.Context, store *receipts.Store) (receipts.SubmissionReceipt, receipts.TransactionReceipt) {
	t.Helper()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-approved-escrow",
		ArtifactLabel:       "artifact/approved-escrow",
		PayloadHash:         "hash-approved-escrow",
		SourceLineageDigest: "lineage-approved-escrow",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "25.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "draft", Amount: "10.00"},
			{Description: "final", Amount: "15.00"},
		},
	})
	require.NoError(t, err)

	return submission, tx
}

func cloneCreateRequest(req escrow.CreateRequest) escrow.CreateRequest {
	cloned := req
	cloned.Amount = new(big.Int).Set(req.Amount)
	if len(req.Milestones) > 0 {
		cloned.Milestones = make([]escrow.MilestoneRequest, len(req.Milestones))
		for i, milestone := range req.Milestones {
			cloned.Milestones[i] = escrow.MilestoneRequest{
				Description: milestone.Description,
				Amount:      new(big.Int).Set(milestone.Amount),
			}
		}
	}
	return cloned
}

func mustBigInt(t *testing.T, value string) *big.Int {
	t.Helper()

	out := new(big.Int)
	_, ok := out.SetString(value, 10)
	require.True(t, ok)
	return out
}
