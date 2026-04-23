package postadjudicationstatus

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/receipts"
)

func TestServiceListCurrentDeadLetters_ReturnsOnlyCurrentDeadLetteredTransactions(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactions = []receipts.TransactionReceipt{
		{
			TransactionReceiptID:           "tx-dead",
			CurrentSubmissionReceiptID:     "sub-dead",
			EscrowAdjudication:             receipts.EscrowAdjudicationRelease,
			EscrowReference:                "escrow-dead",
			CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
			CanonicalSettlementStatus:      receipts.SettlementSettled,
			CanonicalApprovalStatus:        receipts.ApprovalApproved,
			SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
			KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
		},
		{
			TransactionReceiptID:           "tx-retry",
			CurrentSubmissionReceiptID:     "sub-retry",
			EscrowAdjudication:             receipts.EscrowAdjudicationRefund,
			EscrowReference:                "escrow-retry",
			CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
			CanonicalSettlementStatus:      receipts.SettlementSettled,
			CanonicalApprovalStatus:        receipts.ApprovalApproved,
			SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
			KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
		},
		{
			TransactionReceiptID:           "tx-old",
			CurrentSubmissionReceiptID:     "sub-current",
			EscrowAdjudication:             receipts.EscrowAdjudicationRefund,
			EscrowReference:                "escrow-old",
			CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
			CanonicalSettlementStatus:      receipts.SettlementSettled,
			CanonicalApprovalStatus:        receipts.ApprovalApproved,
			SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
			KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
		},
	}
	store.submissions["sub-dead"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:     "sub-dead",
		TransactionReceiptID:    "tx-dead",
		CanonicalApprovalStatus: receipts.ApprovalApproved,
	}
	store.submissions["sub-retry"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:     "sub-retry",
		TransactionReceiptID:    "tx-retry",
		CanonicalApprovalStatus: receipts.ApprovalApproved,
	}
	store.submissions["sub-old"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:     "sub-old",
		TransactionReceiptID:    "tx-old",
		CanonicalApprovalStatus: receipts.ApprovalApproved,
	}
	store.events["sub-dead"] = []receipts.ReceiptEvent{
		{
			SubmissionReceiptID: "sub-dead",
			Source:              "post_adjudication_retry",
			Subtype:             "dead-lettered",
			Type:                receipts.EventSettlementExecutionFailed,
			Reason:              "attempt=4 outcome=release dispatch_reference=dispatch-dead-123 reason=worker exhausted",
		},
	}
	store.events["sub-retry"] = []receipts.ReceiptEvent{
		{
			SubmissionReceiptID: "sub-retry",
			Source:              "post_adjudication_retry",
			Subtype:             "retry-scheduled",
			Type:                receipts.EventSettlementUpdated,
			Reason:              "attempt=2 outcome=refund dispatch_reference=dispatch-retry-123 next_retry_at=2026-04-23T10:00:00Z",
		},
	}
	store.events["sub-old"] = []receipts.ReceiptEvent{
		{
			SubmissionReceiptID: "sub-old",
			Source:              "post_adjudication_retry",
			Subtype:             "dead-lettered",
			Type:                receipts.EventSettlementExecutionFailed,
			Reason:              "attempt=1 outcome=refund dispatch_reference=dispatch-old-123 reason=stale current submission moved",
		},
	}

	svc := NewService(store)

	got, err := svc.ListCurrentDeadLetters(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "tx-dead", got[0].TransactionReceiptID)
	assert.Equal(t, "sub-dead", got[0].SubmissionReceiptID)
	assert.Equal(t, string(receipts.EscrowAdjudicationRelease), got[0].Adjudication)
}

func TestServiceListCurrentDeadLetters_ExtractsFocusedFields(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactions = []receipts.TransactionReceipt{
		{
			TransactionReceiptID:           "tx-1",
			CurrentSubmissionReceiptID:     "sub-1",
			EscrowAdjudication:             receipts.EscrowAdjudicationRefund,
			CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
			CanonicalSettlementStatus:      receipts.SettlementSettled,
			CanonicalApprovalStatus:        receipts.ApprovalApproved,
			SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
			KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
		},
	}
	store.submissions["sub-1"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:     "sub-1",
		TransactionReceiptID:    "tx-1",
		CanonicalApprovalStatus: receipts.ApprovalApproved,
	}
	store.events["sub-1"] = []receipts.ReceiptEvent{
		manualRetryEvent("sub-1", "operator:alice"),
		{
			SubmissionReceiptID: "sub-1",
			Source:              "post_adjudication_retry",
			Subtype:             "retry-scheduled",
			Type:                receipts.EventSettlementUpdated,
			Reason:              "attempt=3 outcome=refund dispatch_reference=dispatch-early-123 next_retry_at=2026-04-23T09:00:00Z",
		},
		{
			SubmissionReceiptID: "sub-1",
			Source:              "post_adjudication_retry",
			Subtype:             "dead-lettered",
			Type:                receipts.EventSettlementExecutionFailed,
			Reason:              "attempt=4 outcome=refund dispatch_reference=dispatch-final-123 dead_lettered_at=2026-04-23T11:45:00Z reason=worker exhausted after 4 attempts",
		},
	}

	svc := NewService(store)

	got, err := svc.ListCurrentDeadLetters(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "tx-1", got[0].TransactionReceiptID)
	assert.Equal(t, "sub-1", got[0].SubmissionReceiptID)
	assert.Equal(t, string(receipts.EscrowAdjudicationRefund), got[0].Adjudication)
	assert.Equal(t, "worker exhausted after 4 attempts", got[0].LatestDeadLetterReason)
	assert.Equal(t, "2026-04-23T11:45:00Z", got[0].LatestDeadLetteredAt)
	assert.Equal(t, "operator:alice", got[0].LatestManualReplayActor)
	assert.Equal(t, 4, got[0].LatestRetryAttempt)
	assert.Equal(t, "dispatch-final-123", got[0].LatestDispatchReference)
}

func TestServiceGetTransactionStatus_ReturnsCanonicalSnapshotAndSummary(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactionByID["tx-1"] = receipts.TransactionReceipt{
		TransactionReceiptID:           "tx-1",
		CurrentSubmissionReceiptID:     "sub-1",
		EscrowAdjudication:             receipts.EscrowAdjudicationRelease,
		EscrowReference:                "escrow-123",
		CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
		CanonicalSettlementStatus:      receipts.SettlementSettled,
		CanonicalApprovalStatus:        receipts.ApprovalApproved,
		SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
		KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
	}
	store.submissions["sub-1"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:     "sub-1",
		TransactionReceiptID:    "tx-1",
		ArtifactLabel:           "artifact",
		PayloadHash:             "payload-hash",
		SourceLineageDigest:     "lineage",
		CanonicalApprovalStatus: receipts.ApprovalApproved,
	}
	store.events["sub-1"] = []receipts.ReceiptEvent{
		manualRetryEvent("sub-1", "operator:alice"),
		{
			SubmissionReceiptID: "sub-1",
			Source:              "post_adjudication_retry",
			Subtype:             "retry-scheduled",
			Type:                receipts.EventSettlementUpdated,
			Reason:              "attempt=2 outcome=release dispatch_reference=dispatch-123 next_retry_at=2026-04-23T09:00:00Z",
		},
		{
			SubmissionReceiptID: "sub-1",
			Source:              "post_adjudication_retry",
			Subtype:             "dead-lettered",
			Type:                receipts.EventSettlementExecutionFailed,
			Reason:              "attempt=3 outcome=release dispatch_reference=dispatch-456 dead_lettered_at=2026-04-23T10:30:00Z reason=terminal worker failure",
		},
	}

	svc := NewService(store)

	got, err := svc.GetTransactionStatus(context.Background(), "tx-1")
	require.NoError(t, err)
	require.Equal(t, store.transactionByID["tx-1"], got.CanonicalSnapshot.TransactionReceipt)
	require.Equal(t, store.submissions["sub-1"], got.CanonicalSnapshot.SubmissionReceipt)
	require.Equal(t, store.events["sub-1"], got.CanonicalSnapshot.SubmissionEvents)
	assert.Equal(t, 3, got.RetryDeadLetterSummary.LatestRetryAttempt)
	assert.Equal(t, "terminal worker failure", got.RetryDeadLetterSummary.LatestDeadLetterReason)
	assert.Equal(t, "2026-04-23T10:30:00Z", got.RetryDeadLetterSummary.LatestDeadLetteredAt)
	assert.Equal(t, "operator:alice", got.RetryDeadLetterSummary.LatestManualReplayActor)
	assert.Equal(t, "dispatch-456", got.RetryDeadLetterSummary.LatestDispatchReference)
}

func TestServiceGetTransactionStatus_ReturnsMissingTransactionFailure(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.getTransactionErr = receipts.ErrTransactionReceiptNotFound
	svc := NewService(store)

	got, err := svc.GetTransactionStatus(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	assert.Equal(t, TransactionStatus{}, got)
}

func TestServiceListCurrentDeadLettersPage_FiltersByOutcomeAttemptAndQuery(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactions = []receipts.TransactionReceipt{
		makeDeadLetterTransaction("tx-a", "sub-a", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-b", "sub-b", receipts.EscrowAdjudicationRefund),
		makeDeadLetterTransaction("tx-c", "sub-c", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-d", "sub-d", receipts.EscrowAdjudicationRelease),
	}
	store.submissions["sub-a"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-a", TransactionReceiptID: "tx-a"}
	store.submissions["sub-b"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-b", TransactionReceiptID: "tx-b"}
	store.submissions["sub-c"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-c", TransactionReceiptID: "tx-c"}
	store.submissions["sub-d"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-d", TransactionReceiptID: "tx-d"}
	store.events["sub-a"] = []receipts.ReceiptEvent{deadLetterEvent("sub-a", 6, "dispatch-a")}
	store.events["sub-b"] = []receipts.ReceiptEvent{deadLetterEvent("sub-b", 5, "dispatch-b")}
	store.events["sub-c"] = []receipts.ReceiptEvent{deadLetterEvent("sub-c", 4, "dispatch-c")}
	store.events["sub-d"] = []receipts.ReceiptEvent{deadLetterEvent("sub-d", 2, "dispatch-d")}

	svc := NewService(store)

	tests := []struct {
		name string
		opts DeadLetterListOptions
		want []string
	}{
		{
			name: "transaction id query",
			opts: DeadLetterListOptions{
				Adjudication:    string(receipts.EscrowAdjudicationRelease),
				RetryAttemptMin: 6,
				RetryAttemptMax: 6,
				Query:           "tx-a",
			},
			want: []string{"tx-a"},
		},
		{
			name: "submission id query",
			opts: DeadLetterListOptions{
				Adjudication:    string(receipts.EscrowAdjudicationRelease),
				RetryAttemptMin: 4,
				RetryAttemptMax: 4,
				Query:           "sub-c",
			},
			want: []string{"tx-c"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := svc.ListCurrentDeadLettersPage(context.Background(), tc.opts)
			require.NoError(t, err)
			require.Equal(t, len(tc.want), got.Count)
			require.Equal(t, len(tc.want), len(got.Items))
			for i, want := range tc.want {
				assert.Equal(t, want, got.Items[i].TransactionReceiptID)
			}
			assert.Equal(t, len(tc.want), got.Total)
			assert.Equal(t, tc.opts.Offset, got.Offset)
			assert.Equal(t, tc.opts.Limit, got.Limit)
		})
	}
}

func TestServiceListCurrentDeadLettersPage_FiltersByManualReplayActorAndDeadLetterWindow(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactions = []receipts.TransactionReceipt{
		makeDeadLetterTransaction("tx-a", "sub-a", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-b", "sub-b", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-c", "sub-c", receipts.EscrowAdjudicationRefund),
	}
	store.submissions["sub-a"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-a", TransactionReceiptID: "tx-a"}
	store.submissions["sub-b"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-b", TransactionReceiptID: "tx-b"}
	store.submissions["sub-c"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-c", TransactionReceiptID: "tx-c"}
	store.events["sub-a"] = []receipts.ReceiptEvent{
		manualRetryEvent("sub-a", "operator:alice"),
		deadLetterEventAt("sub-a", 5, "dispatch-a", "2026-04-23T09:15:00Z"),
	}
	store.events["sub-b"] = []receipts.ReceiptEvent{
		manualRetryEvent("sub-b", "operator:bob"),
		deadLetterEventAt("sub-b", 4, "dispatch-b", "2026-04-23T11:15:00Z"),
	}
	store.events["sub-c"] = []receipts.ReceiptEvent{
		manualRetryEvent("sub-c", "operator:alice"),
		deadLetterEventAt("sub-c", 3, "dispatch-c", "2026-04-23T12:30:00Z"),
	}

	svc := NewService(store)

	got, err := svc.ListCurrentDeadLettersPage(context.Background(), DeadLetterListOptions{
		Adjudication:       string(receipts.EscrowAdjudicationRelease),
		ManualReplayActor:  "operator:bob",
		DeadLetteredAfter:  "2026-04-23T10:00:00Z",
		DeadLetteredBefore: "2026-04-23T12:00:00Z",
	})
	require.NoError(t, err)
	require.Equal(t, 1, got.Total)
	require.Equal(t, 1, got.Count)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "tx-b", got.Items[0].TransactionReceiptID)
	assert.Equal(t, "2026-04-23T11:15:00Z", got.Items[0].LatestDeadLetteredAt)
	assert.Equal(t, "operator:bob", got.Items[0].LatestManualReplayActor)
}

func TestServiceListCurrentDeadLettersPage_ReturnsPaginationMetadata(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactions = []receipts.TransactionReceipt{
		makeDeadLetterTransaction("tx-a", "sub-a", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-b", "sub-b", receipts.EscrowAdjudicationRefund),
		makeDeadLetterTransaction("tx-c", "sub-c", receipts.EscrowAdjudicationRelease),
		makeDeadLetterTransaction("tx-d", "sub-d", receipts.EscrowAdjudicationRelease),
	}
	store.submissions["sub-a"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-a", TransactionReceiptID: "tx-a"}
	store.submissions["sub-b"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-b", TransactionReceiptID: "tx-b"}
	store.submissions["sub-c"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-c", TransactionReceiptID: "tx-c"}
	store.submissions["sub-d"] = receipts.SubmissionReceipt{SubmissionReceiptID: "sub-d", TransactionReceiptID: "tx-d"}
	store.events["sub-a"] = []receipts.ReceiptEvent{deadLetterEvent("sub-a", 6, "dispatch-a")}
	store.events["sub-b"] = []receipts.ReceiptEvent{deadLetterEvent("sub-b", 5, "dispatch-b")}
	store.events["sub-c"] = []receipts.ReceiptEvent{deadLetterEvent("sub-c", 4, "dispatch-c")}
	store.events["sub-d"] = []receipts.ReceiptEvent{deadLetterEvent("sub-d", 2, "dispatch-d")}

	svc := NewService(store)

	got, err := svc.ListCurrentDeadLettersPage(context.Background(), DeadLetterListOptions{
		Offset: 1,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Equal(t, 4, got.Total)
	require.Equal(t, 2, got.Count)
	require.Equal(t, 1, got.Offset)
	require.Equal(t, 2, got.Limit)
	require.Len(t, got.Items, 2)
	assert.Equal(t, "tx-b", got.Items[0].TransactionReceiptID)
	assert.Equal(t, "tx-c", got.Items[1].TransactionReceiptID)
}

func TestServiceGetTransactionStatus_IncludesNavigationHints(t *testing.T) {
	t.Parallel()

	store := newFakeStatusStore()
	store.transactionByID["tx-1"] = makeDeadLetterTransaction("tx-1", "sub-1", receipts.EscrowAdjudicationRelease)
	store.submissions["sub-1"] = receipts.SubmissionReceipt{
		SubmissionReceiptID:  "sub-1",
		TransactionReceiptID: "tx-1",
	}
	store.events["sub-1"] = []receipts.ReceiptEvent{
		deadLetterEvent("sub-1", 3, "dispatch-1"),
	}

	svc := NewService(store)

	got, err := svc.GetTransactionStatus(context.Background(), "tx-1")
	require.NoError(t, err)
	assert.True(t, got.IsDeadLettered)
	assert.True(t, got.CanRetry)
	assert.Equal(t, string(receipts.EscrowAdjudicationRelease), got.Adjudication)
	assert.True(t, got.RetryDeadLetterSummary.HasDeadLetter)
}

type fakeStatusStore struct {
	mu sync.Mutex

	transactions      []receipts.TransactionReceipt
	transactionByID   map[string]receipts.TransactionReceipt
	submissions       map[string]receipts.SubmissionReceipt
	events            map[string][]receipts.ReceiptEvent
	getTransactionErr error
	getSubmissionErr  error
}

func newFakeStatusStore() *fakeStatusStore {
	return &fakeStatusStore{
		transactionByID: make(map[string]receipts.TransactionReceipt),
		submissions:     make(map[string]receipts.SubmissionReceipt),
		events:          make(map[string][]receipts.ReceiptEvent),
	}
}

func (f *fakeStatusStore) ListTransactionReceipts(context.Context) ([]receipts.TransactionReceipt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.transactions == nil {
		return nil, nil
	}
	return append([]receipts.TransactionReceipt(nil), f.transactions...), nil
}

func (f *fakeStatusStore) GetTransactionReceipt(_ context.Context, transactionReceiptID string) (receipts.TransactionReceipt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.getTransactionErr != nil {
		return receipts.TransactionReceipt{}, f.getTransactionErr
	}
	transaction, ok := f.transactionByID[transactionReceiptID]
	if !ok {
		return receipts.TransactionReceipt{}, receipts.ErrTransactionReceiptNotFound
	}
	return transaction, nil
}

func (f *fakeStatusStore) GetSubmissionReceipt(_ context.Context, submissionReceiptID string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.getSubmissionErr != nil {
		return receipts.SubmissionReceipt{}, nil, f.getSubmissionErr
	}
	submission, ok := f.submissions[submissionReceiptID]
	if !ok {
		return receipts.SubmissionReceipt{}, nil, receipts.ErrSubmissionReceiptNotFound
	}
	return submission, append([]receipts.ReceiptEvent(nil), f.events[submissionReceiptID]...), nil
}

func (f *fakeStatusStore) SetTransaction(transaction receipts.TransactionReceipt) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.transactionByID == nil {
		f.transactionByID = make(map[string]receipts.TransactionReceipt)
	}
	f.transactionByID[transaction.TransactionReceiptID] = transaction
}

func (f *fakeStatusStore) SetSubmission(submission receipts.SubmissionReceipt, events []receipts.ReceiptEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.submissions == nil {
		f.submissions = make(map[string]receipts.SubmissionReceipt)
	}
	if f.events == nil {
		f.events = make(map[string][]receipts.ReceiptEvent)
	}
	f.submissions[submission.SubmissionReceiptID] = submission
	f.events[submission.SubmissionReceiptID] = append([]receipts.ReceiptEvent(nil), events...)
}

func (f *fakeStatusStore) SetTransactions(transactions []receipts.TransactionReceipt) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.transactions = append([]receipts.TransactionReceipt(nil), transactions...)
	if f.transactionByID == nil {
		f.transactionByID = make(map[string]receipts.TransactionReceipt)
	}
	for _, transaction := range transactions {
		f.transactionByID[transaction.TransactionReceiptID] = transaction
	}
}

func (f *fakeStatusStore) SetTransactionError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.getTransactionErr = err
}

func (f *fakeStatusStore) SetSubmissionError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.getSubmissionErr = err
}

func (f *fakeStatusStore) TouchSubmission(submissionID string, event receipts.ReceiptEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.events[submissionID] = append(f.events[submissionID], event)
}

func (f *fakeStatusStore) SeedSubmission(submission receipts.SubmissionReceipt) {
	f.SetSubmission(submission, nil)
}

func (f *fakeStatusStore) SeedTransaction(transaction receipts.TransactionReceipt) {
	f.SetTransaction(transaction)
}

func (f *fakeStatusStore) CurrentTransaction(transactionID string) (receipts.TransactionReceipt, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	transaction, ok := f.transactionByID[transactionID]
	return transaction, ok
}

func (f *fakeStatusStore) CurrentSubmission(submissionID string) (receipts.SubmissionReceipt, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	submission, ok := f.submissions[submissionID]
	return submission, ok
}

func (f *fakeStatusStore) CurrentEvents(submissionID string) []receipts.ReceiptEvent {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]receipts.ReceiptEvent(nil), f.events[submissionID]...)
}

var _ receiptStore = (*fakeStatusStore)(nil)

func makeDeadLetterTransaction(transactionID, submissionID string, adjudication receipts.EscrowAdjudicationDecision) receipts.TransactionReceipt {
	return receipts.TransactionReceipt{
		TransactionReceiptID:           transactionID,
		CurrentSubmissionReceiptID:     submissionID,
		EscrowAdjudication:             adjudication,
		CurrentPaymentApprovalStatus:   receipts.PaymentApprovalApproved,
		CanonicalSettlementStatus:      receipts.SettlementSettled,
		CanonicalApprovalStatus:        receipts.ApprovalApproved,
		SettlementProgressionStatus:    receipts.SettlementProgressionDisputeReady,
		KnowledgeExchangeRuntimeStatus: receipts.RuntimeStatusDisputeReady,
	}
}

func deadLetterEvent(submissionID string, attempt int, dispatchReference string) receipts.ReceiptEvent {
	return receipts.ReceiptEvent{
		SubmissionReceiptID: submissionID,
		Source:              "post_adjudication_retry",
		Subtype:             "dead-lettered",
		Type:                receipts.EventSettlementExecutionFailed,
		Reason:              "attempt=" + strconv.Itoa(attempt) + " outcome=release dispatch_reference=" + dispatchReference + " reason=worker exhausted",
	}
}

func deadLetterEventAt(submissionID string, attempt int, dispatchReference string, deadLetteredAt string) receipts.ReceiptEvent {
	return receipts.ReceiptEvent{
		SubmissionReceiptID: submissionID,
		Source:              "post_adjudication_retry",
		Subtype:             "dead-lettered",
		Type:                receipts.EventSettlementExecutionFailed,
		Reason:              "attempt=" + strconv.Itoa(attempt) + " outcome=release dispatch_reference=" + dispatchReference + " dead_lettered_at=" + deadLetteredAt + " reason=worker exhausted",
	}
}

func manualRetryEvent(submissionID string, actor string) receipts.ReceiptEvent {
	return receipts.ReceiptEvent{
		SubmissionReceiptID: submissionID,
		Source:              "post_adjudication_retry",
		Subtype:             "manual-retry-requested",
		Type:                receipts.EventSettlementUpdated,
		Reason:              "actor=" + actor + " reason=manual retry requested",
	}
}

func TestLatestDeadLetterEventParsing(t *testing.T) {
	t.Parallel()

	parsed := parseEventSummary(receipts.ReceiptEvent{
		Reason: "attempt=7 outcome=release dispatch_reference=dispatch-777 reason=worker exhausted again",
	})
	require.Equal(t, 7, parsed.LatestRetryAttempt)
	require.Equal(t, "dispatch-777", parsed.LatestDispatchReference)
	require.Equal(t, "worker exhausted again", parsed.LatestDeadLetterReason)
}

func TestSummaryIsEmptyWithoutDeadLetterEvidence(t *testing.T) {
	t.Parallel()

	summary := summarizeEvents([]receipts.ReceiptEvent{
		{
			Source:  "post_adjudication_retry",
			Subtype: "retry-scheduled",
			Reason:  "attempt=2 outcome=release dispatch_reference=dispatch-1 next_retry_at=2026-04-23T09:00:00Z",
		},
	})
	require.False(t, summary.HasDeadLetter)
	require.Equal(t, 2, summary.LatestRetryAttempt)
	require.Equal(t, "dispatch-1", summary.LatestDispatchReference)
}

func TestSummaryIgnoresUnrelatedEvents(t *testing.T) {
	t.Parallel()

	summary := summarizeEvents([]receipts.ReceiptEvent{
		{
			Source:  "manual",
			Subtype: "note",
			Reason:  "attempt=5 outcome=release dispatch_reference=dispatch-5 reason=ignored",
		},
	})
	require.False(t, summary.HasDeadLetter)
	require.Zero(t, summary.LatestRetryAttempt)
	require.Empty(t, summary.LatestDispatchReference)
	require.Empty(t, summary.LatestDeadLetterReason)
}

func TestSummaryKeepsLastDeadLetterReason(t *testing.T) {
	t.Parallel()

	summary := summarizeEvents([]receipts.ReceiptEvent{
		{
			Source:  "post_adjudication_retry",
			Subtype: "dead-lettered",
			Reason:  "attempt=3 outcome=release dispatch_reference=dispatch-1 reason=first failure",
		},
		{
			Source:  "post_adjudication_retry",
			Subtype: "dead-lettered",
			Reason:  "attempt=4 outcome=release dispatch_reference=dispatch-2 reason=second failure",
		},
	})
	require.True(t, summary.HasDeadLetter)
	require.Equal(t, 4, summary.LatestRetryAttempt)
	require.Equal(t, "dispatch-2", summary.LatestDispatchReference)
	require.Equal(t, "second failure", summary.LatestDeadLetterReason)
}

func TestParseEventSummaryHandlesMissingStructuredData(t *testing.T) {
	t.Parallel()

	parsed := parseEventSummary(receipts.ReceiptEvent{
		Reason: "worker failed",
	})
	require.Equal(t, 0, parsed.LatestRetryAttempt)
	require.Empty(t, parsed.LatestDispatchReference)
	require.Equal(t, "worker failed", parsed.LatestDeadLetterReason)
}

func TestParseEventSummaryUsesDeadLetterReasonSuffix(t *testing.T) {
	t.Parallel()

	parsed := parseEventSummary(receipts.ReceiptEvent{
		Reason: "attempt=9 outcome=refund dispatch_reference=dispatch-9 reason=terminal failure",
	})
	require.Equal(t, "terminal failure", parsed.LatestDeadLetterReason)
}

func TestSummarizeEventsPrefersLatestRelevantEvent(t *testing.T) {
	t.Parallel()

	summary := summarizeEvents([]receipts.ReceiptEvent{
		{
			Source:  "post_adjudication_retry",
			Subtype: "retry-scheduled",
			Reason:  "attempt=1 outcome=refund dispatch_reference=dispatch-1 next_retry_at=2026-04-23T09:00:00Z",
		},
		{
			Source:  "post_adjudication_retry",
			Subtype: "dead-lettered",
			Reason:  "attempt=2 outcome=refund dispatch_reference=dispatch-2 reason=failed permanently",
		},
	})
	require.True(t, summary.HasDeadLetter)
	require.Equal(t, 2, summary.LatestRetryAttempt)
	require.Equal(t, "dispatch-2", summary.LatestDispatchReference)
	require.Equal(t, "failed permanently", summary.LatestDeadLetterReason)
}

func TestDeadLetterTimestampParsingIsStable(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, time.April, 23, 10, 0, 0, 0, time.UTC)
	summary := summarizeEvents([]receipts.ReceiptEvent{
		{
			Source:  "post_adjudication_retry",
			Subtype: "retry-scheduled",
			Reason:  "attempt=5 outcome=release dispatch_reference=dispatch-5 next_retry_at=" + when.Format(time.RFC3339),
		},
	})
	require.Equal(t, 5, summary.LatestRetryAttempt)
	require.Equal(t, "dispatch-5", summary.LatestDispatchReference)
}

var _ interface {
	ListTransactionReceipts(context.Context) ([]receipts.TransactionReceipt, error)
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
} = (*fakeStatusStore)(nil)
