package receipts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	return NewStore()
}

func TestCreateSubmissionReceipt_CreatesTransactionAndCurrentPointer(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-1",
		ArtifactLabel:       "research-memo-v1",
		PayloadHash:         "hash-1",
		SourceLineageDigest: "lineage-1",
	})
	require.NoError(t, err)
	require.NotEmpty(t, sub.SubmissionReceiptID)
	require.NotEmpty(t, tx.TransactionReceiptID)
	require.Equal(t, sub.SubmissionReceiptID, tx.CurrentSubmissionReceiptID)
	require.Equal(t, ApprovalPending, sub.CanonicalApprovalStatus)
	require.Equal(t, SettlementPending, tx.CanonicalSettlementStatus)
	require.Equal(t, PaymentApprovalPending, tx.CurrentPaymentApprovalStatus)
	require.Empty(t, tx.CanonicalDecision)
	require.Empty(t, tx.CanonicalSettlementHint)
}

func TestCreateSubmissionReceipt_RejectsEmptyInput(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{})
	require.ErrorIs(t, err, ErrInvalidSubmissionInput)
}

func TestOpenKnowledgeExchangeTransaction_BindsCanonicalInputs(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)
	require.Equal(t, "did:lango:peer-1", tx.Counterparty)
	require.Equal(t, "artifact/research-note", tx.RequestedScope)
	require.Equal(t, "quote:0.50-usdc", tx.PriceContext)
	require.Equal(t, "trust:0.72", tx.TrustContext)
	require.Equal(t, RuntimeStatusOpened, tx.KnowledgeExchangeRuntimeStatus)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, stored.TransactionReceiptID)
	require.Equal(t, "deal-open-1", stored.TransactionID)
	require.Equal(t, "did:lango:peer-1", stored.Counterparty)
	require.Equal(t, "artifact/research-note", stored.RequestedScope)
	require.Equal(t, "quote:0.50-usdc", stored.PriceContext)
	require.Equal(t, "trust:0.72", stored.TrustContext)
	require.Equal(t, RuntimeStatusOpened, stored.KnowledgeExchangeRuntimeStatus)
}

func TestApplyKnowledgeExchangeRuntimeProgression_RejectsIllegalBranchRewinds(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/code-review",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.83",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, "")
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusOpened, "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidKnowledgeExchangeRuntimeState)
}

func TestOpenKnowledgeExchangeTransaction_RebindsRuntimeStateToOpened(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-3",
		Counterparty:   "did:lango:peer-3",
		RequestedScope: "artifact/runtime-reset",
		PriceContext:   "quote:2.00-usdc",
		TrustContext:   "trust:0.91",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, "")
	require.NoError(t, err)

	reopened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-3",
		Counterparty:   "did:lango:peer-3",
		RequestedScope: "artifact/runtime-reset",
		PriceContext:   "quote:2.00-usdc",
		TrustContext:   "trust:0.91",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, reopened.TransactionReceiptID)
	require.Equal(t, RuntimeStatusOpened, reopened.KnowledgeExchangeRuntimeStatus)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, RuntimeStatusOpened, stored.KnowledgeExchangeRuntimeStatus)
}

func TestOpenKnowledgeExchangeTransaction_RejectsConflictingCanonicalInputs(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-4",
		Counterparty:   "did:lango:peer-4",
		RequestedScope: "artifact/baseline",
		PriceContext:   "quote:3.00-usdc",
		TrustContext:   "trust:0.66",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, "")
	require.NoError(t, err)

	_, err = store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-4",
		Counterparty:   "did:lango:peer-conflict",
		RequestedScope: "artifact/baseline",
		PriceContext:   "quote:3.00-usdc",
		TrustContext:   "trust:0.66",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSubmissionInput)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "did:lango:peer-4", stored.Counterparty)
	require.Equal(t, RuntimeStatusPaymentApproved, stored.KnowledgeExchangeRuntimeStatus)
}

func TestApplySettlementProgression_MapsReleaseOutcomeToCanonicalState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-settle-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	updated, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, "approve", "")
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
	require.Equal(t, "approve", updated.SettlementProgressionReason)
}

func TestApplySettlementProgression_RejectsIllegalRewind(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-settle-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/code-review",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.83",
	})
	require.NoError(t, err)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, "approve", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionPending, "rewind", "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestApplyKnowledgeExchangeRuntimeProgression_RejectsNonexistentSubmissionPointer(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-5",
		Counterparty:   "did:lango:peer-5",
		RequestedScope: "artifact/submission-check",
		PriceContext:   "quote:4.00-usdc",
		TrustContext:   "trust:0.77",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, "missing-submission")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestApplyKnowledgeExchangeRuntimeProgression_RejectsForeignSubmissionPointer(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-6",
		Counterparty:   "did:lango:peer-6",
		RequestedScope: "artifact/runtime-owner",
		PriceContext:   "quote:5.00-usdc",
		TrustContext:   "trust:0.81",
	})
	require.NoError(t, err)

	foreignSub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-runtime-foreign",
		ArtifactLabel:       "artifact/runtime-foreign",
		PayloadHash:         "hash-runtime-foreign",
		SourceLineageDigest: "lineage-runtime-foreign",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, foreignSub.SubmissionReceiptID)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestCreateSubmissionReceipt_UpdatesCurrentPointerOnSecondSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-3",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-3",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.NotEqual(t, first.SubmissionReceiptID, second.SubmissionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	got, events, err := store.GetSubmissionReceipt(ctx, second.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "memo-b", got.ArtifactLabel)
	require.Empty(t, events)
}

func TestCreateSubmissionReceipt_ResetsEscrowExecutionMetadataOnCurrentSubmissionChange(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-reset",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "5.00",
		Reason:    "knowledge exchange",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "2.00"},
		},
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-reset",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.NotEqual(t, first.SubmissionReceiptID, second.SubmissionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)
	require.Empty(t, nextTx.EscrowExecutionStatus)
	require.Empty(t, nextTx.EscrowReference)
	require.Nil(t, nextTx.EscrowExecutionInput)

	gotTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, second.SubmissionReceiptID, gotTx.CurrentSubmissionReceiptID)
	require.Empty(t, gotTx.EscrowExecutionStatus)
	require.Empty(t, gotTx.EscrowReference)
	require.Nil(t, gotTx.EscrowExecutionInput)
}

func TestAppendReceiptEvent_PreservesCanonicalReceiptAndTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-2",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-2",
		SourceLineageDigest: "lineage-2",
	})
	require.NoError(t, err)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, EventApprovalRequested)
	require.NoError(t, err)

	got, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, ApprovalPending, got.CanonicalApprovalStatus)
	require.Len(t, events, 1)
	require.Equal(t, EventApprovalRequested, events[0].Type)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, "manual", events[0].Source)
	require.Equal(t, string(EventApprovalRequested), events[0].Subtype)
}

func TestApplyUpfrontPaymentApproval_ApprovesAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-approval",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-approval",
		SourceLineageDigest: "lineage-approval",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, updatedTx.TransactionReceiptID)
	require.Equal(t, PaymentApprovalApproved, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "approve", updatedTx.CanonicalDecision)
	require.Equal(t, "prepay", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, EventPaymentApproval, events[0].Type)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, "approval", events[0].Source)
	require.Equal(t, "approval.upfront_payment", events[0].Subtype)
}

func TestApplyUpfrontPaymentApproval_RejectsAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-reject",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-reject",
		SourceLineageDigest: "lineage-reject",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionReject,
		Reason:        "Amount exceeds max prepay policy.",
		SuggestedMode: paymentapproval.ModeReject,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalRejected, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "reject", updatedTx.CanonicalDecision)
	require.Equal(t, "reject", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, EventPaymentApproval, events[0].Type)
}

func TestApplyUpfrontPaymentApproval_EscalatesAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escalate",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-escalate",
		SourceLineageDigest: "lineage-escalate",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionEscalate,
		Reason:        "Trust score is in an edge-case range for upfront payment.",
		SuggestedMode: paymentapproval.ModeEscalate,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalEscalated, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "escalate", updatedTx.CanonicalDecision)
	require.Equal(t, "escalate", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, EventPaymentApproval, events[0].Type)
}

func TestApplyUpfrontPaymentApproval_RejectsInvalidTransactionOrSubmissionIdentifiers(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-ids",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-ids",
		SourceLineageDigest: "lineage-ids",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	_, err = store.ApplyUpfrontPaymentApproval(ctx, "missing-transaction", sub.SubmissionReceiptID, outcome)
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, "missing-submission", outcome)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	otherSub, otherTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-other",
		ArtifactLabel:       "memo-other",
		PayloadHash:         "hash-other",
		SourceLineageDigest: "lineage-other",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, otherSub.SubmissionReceiptID, outcome)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, otherTx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestApplyUpfrontPaymentApproval_RejectsInvalidOutcomeWithoutMutatingStorage(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-invalid-outcome",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-invalid-outcome",
		SourceLineageDigest: "lineage-invalid-outcome",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.Decision("bogus"),
		Reason:        "Impossible state.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPaymentApprovalStatus)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, events)
	require.Equal(t, ApprovalPending, gotSub.CanonicalApprovalStatus)

	updatedTx := store.transactions[tx.TransactionReceiptID]
	require.Equal(t, PaymentApprovalPending, updatedTx.CurrentPaymentApprovalStatus)
	require.Empty(t, updatedTx.CanonicalDecision)
	require.Empty(t, updatedTx.CanonicalSettlementHint)
}

func TestApplyUpfrontPaymentApproval_UsesExplicitSubmissionInMultiSubmissionTransaction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-multi",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-multi",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalApproved, updatedTx.CurrentPaymentApprovalStatus)

	_, firstEvents, err := store.GetSubmissionReceipt(ctx, first.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, firstEvents, 1)
	require.Equal(t, first.SubmissionReceiptID, firstEvents[0].SubmissionReceiptID)

	_, secondEvents, err := store.GetSubmissionReceipt(ctx, second.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, secondEvents)
}

func TestBindEscrowExecutionInput_PersistsCanonicalInputOnTransaction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	milestones := []EscrowMilestoneInput{
		{Description: "draft", Amount: "1.50"},
		{Description: "final", Amount: "2.00"},
	}
	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-bind",
		ArtifactLabel:       "artifact/escrow-bind",
		PayloadHash:         "hash-escrow-bind",
		SourceLineageDigest: "lineage-escrow-bind",
	})
	require.NoError(t, err)

	updated, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:   "did:lango:buyer",
		SellerDID:  "did:lango:seller",
		Amount:     "3.50",
		Reason:     "knowledge exchange",
		TaskID:     "task-escrow-bind",
		Milestones: milestones,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.EscrowExecutionInput)
	require.Equal(t, EscrowExecutionStatusPending, updated.EscrowExecutionStatus)
	require.Equal(t, "did:lango:buyer", updated.EscrowExecutionInput.BuyerDID)
	require.Equal(t, "3.50", updated.EscrowExecutionInput.Amount)
	require.Len(t, updated.EscrowExecutionInput.Milestones, 2)

	milestones[0].Amount = "9.99"
	updated.EscrowExecutionInput.Milestones[0].Amount = "7.77"

	gotTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.NotNil(t, gotTx.EscrowExecutionInput)
	require.Equal(t, "1.50", gotTx.EscrowExecutionInput.Milestones[0].Amount)
	require.Equal(t, "2.00", gotTx.EscrowExecutionInput.Milestones[1].Amount)

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestBindEscrowExecutionInput_ResetsEscrowReferenceOnRebind(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-rebind",
		ArtifactLabel:       "artifact/escrow-rebind",
		PayloadHash:         "hash-escrow-rebind",
		SourceLineageDigest: "lineage-escrow-rebind",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "6.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)

	updated, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "6.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusPending, updated.EscrowExecutionStatus)
	require.Empty(t, updated.EscrowReference)

	gotTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusPending, gotTx.EscrowExecutionStatus)
	require.Empty(t, gotTx.EscrowReference)
}

func TestBindEscrowExecutionInput_RejectsNonCurrentSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-stale-bind",
		ArtifactLabel:       "artifact/escrow-stale-bind-1",
		PayloadHash:         "hash-escrow-stale-bind-1",
		SourceLineageDigest: "lineage-escrow-stale-bind-1",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-stale-bind",
		ArtifactLabel:       "artifact/escrow-stale-bind-2",
		PayloadHash:         "hash-escrow-stale-bind-2",
		SourceLineageDigest: "lineage-escrow-stale-bind-2",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "6.00",
		Reason:    "knowledge exchange",
	})
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)

	gotTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Nil(t, gotTx.EscrowExecutionInput)
	require.Empty(t, gotTx.EscrowExecutionStatus)
}

func TestApplyEscrowExecutionProgress_RecordsCreatedFundedAndFailed(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-progress",
		ArtifactLabel:       "artifact/escrow-progress",
		PayloadHash:         "hash-escrow-progress",
		SourceLineageDigest: "lineage-escrow-progress",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow-progress",
		Milestones: []EscrowMilestoneInput{
			{Description: "delivery", Amount: "4.00"},
		},
	})
	require.NoError(t, err)

	started, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusPending, started.EscrowExecutionStatus)
	require.Empty(t, started.EscrowReference)

	created, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusCreated, created.EscrowExecutionStatus)
	require.Empty(t, created.EscrowReference)

	updated, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-1", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, "escrow-1", updated.EscrowReference)

	failed, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFailed, "escrow-1", EventEscrowExecutionFailed, "funding reverted")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFailed, failed.EscrowExecutionStatus)
	require.Equal(t, "escrow-1", failed.EscrowReference)

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 4)
	require.Equal(t, EventEscrowExecutionStarted, events[0].Type)
	require.Equal(t, EventEscrowExecutionCreated, events[1].Type)
	require.Equal(t, EventEscrowExecutionFunded, events[2].Type)
	require.Equal(t, EventEscrowExecutionFailed, events[3].Type)
	require.Equal(t, "funding reverted", events[3].Reason)
}

func TestApplyEscrowExecutionProgress_RejectsInvalidStatusAndUnboundInput(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-invalid-progress",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-invalid-progress",
		SourceLineageDigest: "lineage-invalid-progress",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatus("bogus"), "", EventEscrowExecutionCreated, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionStatus)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "", EventEscrowExecutionFunded, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "", EventEscrowExecutionCreated, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func TestApplyEscrowExecutionProgress_RejectsNonCurrentSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-stale-progress",
		ArtifactLabel:       "artifact/escrow-stale-progress-1",
		PayloadHash:         "hash-escrow-stale-progress-1",
		SourceLineageDigest: "lineage-escrow-stale-progress-1",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-stale-progress",
		ArtifactLabel:       "artifact/escrow-stale-progress-2",
		PayloadHash:         "hash-escrow-stale-progress-2",
		SourceLineageDigest: "lineage-escrow-stale-progress-2",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, second.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func TestApplyEscrowExecutionProgress_RejectsIllegalTransitions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-illegal-transitions",
		ArtifactLabel:       "artifact/escrow-illegal-transitions",
		PayloadHash:         "hash-escrow-illegal-transitions",
		SourceLineageDigest: "lineage-escrow-illegal-transitions",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "escrow-created", EventEscrowExecutionCreated, "")
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFailed, "escrow-created", EventEscrowExecutionFailed, "fund failed")
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "escrow-created-2", EventEscrowExecutionCreated, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func TestApplyEscrowExecutionProgress_RejectsTransitionsFromFunded(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-funded-transitions",
		ArtifactLabel:       "artifact/escrow-funded-transitions",
		PayloadHash:         "hash-escrow-funded-transitions",
		SourceLineageDigest: "lineage-escrow-funded-transitions",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "escrow-funded", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-funded", EventEscrowExecutionFunded, "")
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "escrow-funded", EventEscrowExecutionCreated, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func TestAppendReceiptEvent_RejectsInvalidEventType(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-4",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-4",
		SourceLineageDigest: "lineage-4",
	})
	require.NoError(t, err)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, "")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, EventType("unknown"))
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)
}

func TestAppendReceiptEvent_AllowsEscrowExecutionEventTypes(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-events",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-escrow-events",
		SourceLineageDigest: "lineage-escrow-events",
	})
	require.NoError(t, err)

	for _, eventType := range []EventType{
		EventEscrowExecutionStarted,
		EventEscrowExecutionCreated,
		EventEscrowExecutionFunded,
		EventEscrowExecutionFailed,
	} {
		err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, eventType)
		require.NoError(t, err)
	}
}

func TestAppendReceiptEvent_RejectsMissingSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AppendReceiptEvent(ctx, "missing-submission", EventApprovalRequested)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}
