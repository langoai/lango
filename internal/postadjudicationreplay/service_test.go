package postadjudicationreplay

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/receipts"
)

func replayPolicy() ReplayPolicy {
	return ReplayPolicyFromConfig(&config.Config{
		Replay: config.ReplayConfig{
			AllowedActors:        []string{"operator:alice", "operator:bob"},
			ReleaseAllowedActors: []string{"operator:alice"},
			RefundAllowedActors:  []string{"operator:alice", "operator:bob"},
		},
	})
}

func TestServiceReplay_DeniesWhenActorIsUnresolved(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	dispatcher := &fakeReplayDispatcher{}
	svc := NewService(store, dispatcher, replayPolicy())

	result, err := svc.Replay(context.Background(), Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrActorUnresolved)
	require.Equal(t, Result{}, result)
	assert.Equal(t, 0, store.manualRetryCalls)
	assert.Equal(t, 0, dispatcher.calls)
}

func TestExecutionPolicy_DefaultsToManualRecoveryWhenFlagsAreAbsent(t *testing.T) {
	t.Parallel()

	mode, err := DefaultExecutionPolicy().Resolve(false, false)

	require.NoError(t, err)
	assert.Equal(t, ExecutionModeManualRecovery, mode)
}

func TestExecutionPolicy_ExplicitFlagsOverrideDefault(t *testing.T) {
	t.Parallel()

	policy := ExecutionPolicy{DefaultMode: ExecutionModeBackground}

	mode, err := policy.Resolve(true, false)
	require.NoError(t, err)
	assert.Equal(t, ExecutionModeInline, mode)

	mode, err = policy.Resolve(false, true)
	require.NoError(t, err)
	assert.Equal(t, ExecutionModeBackground, mode)
}

func TestExecutionPolicy_RejectsConflictingFlags(t *testing.T) {
	t.Parallel()

	_, err := DefaultExecutionPolicy().Resolve(true, true)

	require.Error(t, err)
	assert.ErrorContains(t, err, "mutually exclusive")
}

func TestRecoveryPolicy_RecognizesCanonicalDeadLetterEvidence(t *testing.T) {
	t.Parallel()

	policy := DefaultRecoveryPolicy()
	events := []receipts.ReceiptEvent{
		{
			Source:  receipts.PostAdjudicationRecoveryEventSource,
			Subtype: receipts.PostAdjudicationManualRetryRequestedSubtype,
		},
		{
			Source:  receipts.PostAdjudicationRecoveryEventSource,
			Subtype: receipts.PostAdjudicationDeadLetteredSubtype,
		},
	}

	assert.True(t, policy.HasDeadLetterEvidence(events))
	assert.False(t, policy.HasDeadLetterEvidence(events[:1]))
}

func TestRetryKeyHelpers_RoundTripPromptAndIdentity(t *testing.T) {
	t.Parallel()

	prompt := BuildBackgroundDispatchPrompt(
		receipts.EscrowAdjudicationRefund,
		"tx-123",
		"sub-456",
		"escrow-789",
	)

	retryKey := RetryKeyFromPrompt(prompt)
	assert.Equal(t, CanonicalRetryKey("tx-123", receipts.EscrowAdjudicationRefund), retryKey)

	transactionReceiptID, outcome, ok := ParseRetryKey(retryKey)
	require.True(t, ok)
	assert.Equal(t, "tx-123", transactionReceiptID)
	assert.Equal(t, receipts.EscrowAdjudicationRefund, outcome)
}

func TestServiceReplay_DeniesWhenActorIsResolvedButNotAllowed(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	dispatcher := &fakeReplayDispatcher{}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:bob")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrReplayNotAllowed)
	require.Equal(t, Result{}, result)
	assert.Equal(t, 0, store.manualRetryCalls)
	assert.Equal(t, 0, dispatcher.calls)
}

func TestServiceReplay_AllowsActorForReleaseOutcome(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	dispatcher := &fakeReplayDispatcher{
		receipt: BackgroundDispatchReceipt{
			Status:               "queued",
			DispatchReference:    "dispatch-123",
			TransactionReceiptID: store.transaction.TransactionReceiptID,
			SubmissionReceiptID:  store.submission.SubmissionReceiptID,
			EscrowReference:      store.transaction.EscrowReference,
			Outcome:              string(store.transaction.EscrowAdjudication),
		},
	}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.NoError(t, err)
	require.Equal(t, store.transaction, result.CanonicalAdjudication.TransactionReceipt)
	require.Equal(t, store.submission, result.CanonicalAdjudication.SubmissionReceipt)
	require.Equal(t, store.events, result.CanonicalAdjudication.SubmissionEvents)
	require.NotNil(t, result.BackgroundDispatchReceipt)
	assert.Equal(t, "queued", result.BackgroundDispatchReceipt.Status)
	assert.Equal(t, "dispatch-123", result.BackgroundDispatchReceipt.DispatchReference)
	assert.Equal(t, store.transaction.TransactionReceiptID, result.BackgroundDispatchReceipt.TransactionReceiptID)
	assert.Equal(t, store.submission.SubmissionReceiptID, result.BackgroundDispatchReceipt.SubmissionReceiptID)
	assert.Equal(t, store.transaction.EscrowReference, result.BackgroundDispatchReceipt.EscrowReference)
	assert.Equal(t, string(store.transaction.EscrowAdjudication), result.BackgroundDispatchReceipt.Outcome)
	assert.Equal(t, 1, store.manualRetryCalls)
	assert.Equal(t, 1, dispatcher.calls)
	assert.True(t, strings.Contains(dispatcher.lastRequest.Prompt, "release_escrow_settlement"))
	assert.True(t, strings.Contains(dispatcher.lastRequest.Prompt, store.transaction.TransactionReceiptID))
}

func TestServiceReplay_AllowsActorForRefundOutcome(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	store.transaction.EscrowAdjudication = receipts.EscrowAdjudicationRefund
	dispatcher := &fakeReplayDispatcher{
		receipt: BackgroundDispatchReceipt{
			Status:               "queued",
			DispatchReference:    "dispatch-456",
			TransactionReceiptID: store.transaction.TransactionReceiptID,
			SubmissionReceiptID:  store.submission.SubmissionReceiptID,
			EscrowReference:      store.transaction.EscrowReference,
			Outcome:              string(store.transaction.EscrowAdjudication),
		},
	}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:bob")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.NoError(t, err)
	require.Equal(t, receipts.EscrowAdjudicationRefund, result.CanonicalAdjudication.TransactionReceipt.EscrowAdjudication)
	require.NotNil(t, result.BackgroundDispatchReceipt)
	assert.Equal(t, "dispatch-456", result.BackgroundDispatchReceipt.DispatchReference)
	assert.Equal(t, string(receipts.EscrowAdjudicationRefund), result.BackgroundDispatchReceipt.Outcome)
	assert.Equal(t, 1, store.manualRetryCalls)
	assert.Equal(t, 1, dispatcher.calls)
	assert.True(t, strings.Contains(dispatcher.lastRequest.Prompt, "refund_escrow_settlement"))
}

func TestServiceReplay_DeniesMissingTransactionReceipt(t *testing.T) {
	t.Parallel()

	store := &fakeReplayStore{getTransactionErr: receipts.ErrTransactionReceiptNotFound}
	dispatcher := &fakeReplayDispatcher{}
	svc := NewService(store, dispatcher, replayPolicy())

	result, err := svc.Replay(context.Background(), Request{TransactionReceiptID: "missing"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	require.Equal(t, Result{}, result)
	assert.Equal(t, 0, store.manualRetryCalls)
	assert.Equal(t, 0, dispatcher.calls)
}

func TestServiceReplay_DeniesWhenDeadLetterEvidenceIsMissing(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	store.events = nil
	dispatcher := &fakeReplayDispatcher{}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDeadLetterEvidenceMissing)
	require.Equal(t, Result{}, result)
	assert.Equal(t, 0, store.manualRetryCalls)
	assert.Equal(t, 0, dispatcher.calls)
}

func TestServiceReplay_DeniesWhenCanonicalAdjudicationIsMissing(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	store.transaction.EscrowAdjudication = ""
	store.events = []receipts.ReceiptEvent{
		{
			SubmissionReceiptID: store.submission.SubmissionReceiptID,
			Source:              receipts.PostAdjudicationRecoveryEventSource,
			Subtype:             receipts.PostAdjudicationDeadLetteredSubtype,
			Type:                receipts.EventSettlementExecutionFailed,
			Reason:              "attempt=3 outcome=release reason=timeout",
		},
	}
	dispatcher := &fakeReplayDispatcher{}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCanonicalAdjudicationMissing)
	require.Equal(t, Result{}, result)
	assert.Equal(t, 0, store.manualRetryCalls)
	assert.Equal(t, 0, dispatcher.calls)
}

func TestServiceReplay_SuccessReturnsCanonicalAdjudicationSnapshotAndDispatchReceipt(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	dispatcher := &fakeReplayDispatcher{
		receipt: BackgroundDispatchReceipt{
			Status:               "queued",
			DispatchReference:    "dispatch-123",
			TransactionReceiptID: store.transaction.TransactionReceiptID,
			SubmissionReceiptID:  store.submission.SubmissionReceiptID,
			EscrowReference:      store.transaction.EscrowReference,
			Outcome:              string(store.transaction.EscrowAdjudication),
		},
	}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.NoError(t, err)
	require.Equal(t, store.transaction, result.CanonicalAdjudication.TransactionReceipt)
	require.Equal(t, store.submission, result.CanonicalAdjudication.SubmissionReceipt)
	require.Equal(t, store.events, result.CanonicalAdjudication.SubmissionEvents)
	require.NotNil(t, result.BackgroundDispatchReceipt)
	assert.Equal(t, "queued", result.BackgroundDispatchReceipt.Status)
	assert.Equal(t, "dispatch-123", result.BackgroundDispatchReceipt.DispatchReference)
	assert.Equal(t, store.transaction.TransactionReceiptID, result.BackgroundDispatchReceipt.TransactionReceiptID)
	assert.Equal(t, store.submission.SubmissionReceiptID, result.BackgroundDispatchReceipt.SubmissionReceiptID)
	assert.Equal(t, store.transaction.EscrowReference, result.BackgroundDispatchReceipt.EscrowReference)
	assert.Equal(t, string(store.transaction.EscrowAdjudication), result.BackgroundDispatchReceipt.Outcome)
	assert.Equal(t, 1, store.manualRetryCalls)
	assert.Equal(t, 1, dispatcher.calls)
	assert.True(t, strings.Contains(dispatcher.lastRequest.Prompt, "release_escrow_settlement"))
	assert.True(t, strings.Contains(dispatcher.lastRequest.Prompt, store.transaction.TransactionReceiptID))
}

func TestServiceReplay_FailureKeepsCanonicalStateUnchanged(t *testing.T) {
	t.Parallel()

	store := newReplayStore()
	original := store.transaction
	dispatcher := &fakeReplayDispatcher{err: errors.New("dispatch failed")}
	svc := NewService(store, dispatcher, replayPolicy())

	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:alice")
	result, err := svc.Replay(ctx, Request{TransactionReceiptID: store.transaction.TransactionReceiptID})

	require.Error(t, err)
	assert.ErrorContains(t, err, "dispatch failed")
	require.Equal(t, original, store.transaction)
	require.Equal(t, Result{
		CanonicalAdjudication: CanonicalAdjudicationSnapshot{
			TransactionReceipt: original,
			SubmissionReceipt:  store.submission,
			SubmissionEvents:   store.events,
		},
	}, result)
	assert.Equal(t, 1, store.manualRetryCalls)
	assert.Equal(t, 1, dispatcher.calls)
	assert.Nil(t, result.BackgroundDispatchReceipt)
}

type fakeReplayStore struct {
	mu sync.Mutex

	transaction receipts.TransactionReceipt
	submission  receipts.SubmissionReceipt
	events      []receipts.ReceiptEvent

	getTransactionErr error
	getSubmissionErr  error
	recordRetryErr    error

	manualRetryCalls int
	lastRetry        receipts.ManualRetryRequestedRequest
}

func newReplayStore() *fakeReplayStore {
	return &fakeReplayStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
			EscrowAdjudication:          receipts.EscrowAdjudicationRelease,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		events: []receipts.ReceiptEvent{
			{
				SubmissionReceiptID: "sub-1",
				Source:              receipts.PostAdjudicationRecoveryEventSource,
				Subtype:             receipts.PostAdjudicationDeadLetteredSubtype,
				Type:                receipts.EventSettlementExecutionFailed,
				Reason:              "attempt=3 outcome=release reason=timeout",
			},
		},
	}
}

func (f *fakeReplayStore) GetTransactionReceipt(_ context.Context, transactionReceiptID string) (receipts.TransactionReceipt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.getTransactionErr != nil {
		return receipts.TransactionReceipt{}, f.getTransactionErr
	}
	if transactionReceiptID != f.transaction.TransactionReceiptID {
		return receipts.TransactionReceipt{}, receipts.ErrTransactionReceiptNotFound
	}
	return f.transaction, nil
}

func (f *fakeReplayStore) GetSubmissionReceipt(_ context.Context, submissionReceiptID string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.getSubmissionErr != nil {
		return receipts.SubmissionReceipt{}, nil, f.getSubmissionErr
	}
	if submissionReceiptID != f.submission.SubmissionReceiptID {
		return receipts.SubmissionReceipt{}, nil, receipts.ErrSubmissionReceiptNotFound
	}
	return f.submission, append([]receipts.ReceiptEvent(nil), f.events...), nil
}

func (f *fakeReplayStore) RecordManualRetryRequested(ctx context.Context, req receipts.ManualRetryRequestedRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.manualRetryCalls++
	f.lastRetry = req
	if f.recordRetryErr != nil {
		return f.recordRetryErr
	}
	reason := req.Reason
	if actor := strings.TrimSpace(ctxkeys.PrincipalFromContext(ctx)); actor != "" {
		reason = "actor=" + actor + " manual_replay_at=2026-04-24T00:00:00Z reason=" + reason
	}
	f.events = append(f.events, receipts.ReceiptEvent{
		SubmissionReceiptID: f.submission.SubmissionReceiptID,
		Source:              receipts.PostAdjudicationRecoveryEventSource,
		Subtype:             receipts.PostAdjudicationManualRetryRequestedSubtype,
		Type:                receipts.EventSettlementUpdated,
		Reason:              reason,
	})
	return nil
}

type fakeReplayDispatcher struct {
	mu sync.Mutex

	calls       int
	lastRequest BackgroundDispatchRequest
	receipt     BackgroundDispatchReceipt
	err         error
}

func (f *fakeReplayDispatcher) Dispatch(_ context.Context, req BackgroundDispatchRequest) (BackgroundDispatchReceipt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	f.lastRequest = req
	if f.err != nil {
		return BackgroundDispatchReceipt{}, f.err
	}
	if f.receipt.Status == "" {
		f.receipt = BackgroundDispatchReceipt{
			Status:               "queued",
			DispatchReference:    "dispatch-123",
			TransactionReceiptID: req.TransactionReceiptID,
			SubmissionReceiptID:  req.SubmissionReceiptID,
			EscrowReference:      req.EscrowReference,
			Outcome:              string(req.Outcome),
		}
	}
	return f.receipt, nil
}
