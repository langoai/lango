package receipts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	return NewStore()
}

func createSubmittedTransaction(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()

	submission, transaction, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       transactionID,
		ArtifactLabel:       "artifact-" + transactionID,
		PayloadHash:         "hash-" + transactionID,
		SourceLineageDigest: "lineage-" + transactionID,
	})
	require.NoError(t, err)

	return submission, transaction
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

func TestCreateSubmissionReceipt_RejectsEachRequiredField(t *testing.T) {
	t.Parallel()

	valid := CreateSubmissionInput{
		TransactionID:       "tx-required",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	}
	tests := []struct {
		name string
		edit func(*CreateSubmissionInput)
	}{
		{name: "transaction_id", edit: func(in *CreateSubmissionInput) { in.TransactionID = "" }},
		{name: "artifact_label", edit: func(in *CreateSubmissionInput) { in.ArtifactLabel = "" }},
		{name: "payload_hash", edit: func(in *CreateSubmissionInput) { in.PayloadHash = "" }},
		{name: "source_lineage_digest", edit: func(in *CreateSubmissionInput) { in.SourceLineageDigest = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := valid
			tt.edit(&in)
			require.ErrorIs(t, validateCreateSubmissionInput(in), ErrInvalidSubmissionInput)
		})
	}
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

func TestKnowledgeExchangeRuntimeTransition_AllowsForwardBranches(t *testing.T) {
	t.Parallel()

	allowed := [][2]KnowledgeExchangeRuntimeStatus{
		{"", RuntimeStatusOpened},
		{RuntimeStatusOpened, RuntimeStatusExportabilityAdvisory},
		{RuntimeStatusOpened, RuntimeStatusPaymentApproved},
		{RuntimeStatusExportabilityAdvisory, RuntimeStatusPaymentApproved},
		{RuntimeStatusPaymentApproved, RuntimeStatusPaymentAuthorized},
		{RuntimeStatusPaymentApproved, RuntimeStatusEscrowFunded},
		{RuntimeStatusPaymentAuthorized, RuntimeStatusWorkStarted},
		{RuntimeStatusEscrowFunded, RuntimeStatusWorkStarted},
		{RuntimeStatusWorkStarted, RuntimeStatusSubmissionReceived},
		{RuntimeStatusSubmissionReceived, RuntimeStatusReleaseApproved},
		{RuntimeStatusSubmissionReceived, RuntimeStatusRevisionRequested},
		{RuntimeStatusSubmissionReceived, RuntimeStatusEscalated},
		{RuntimeStatusSubmissionReceived, RuntimeStatusDisputeReady},
		{RuntimeStatusRevisionRequested, RuntimeStatusSubmissionReceived},
	}
	for _, transition := range allowed {
		require.NoError(t, validateKnowledgeExchangeRuntimeTransition(transition[0], transition[1]))
	}
	require.ErrorIs(t, validateKnowledgeExchangeRuntimeTransition(RuntimeStatusReleaseApproved, RuntimeStatusOpened), ErrInvalidKnowledgeExchangeRuntimeState)
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

func TestOpenKnowledgeExchangeTransaction_RejectsAllCanonicalConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*OpenTransactionInput)
	}{
		{name: "counterparty", edit: func(in *OpenTransactionInput) { in.Counterparty = "did:lango:other" }},
		{name: "requested_scope", edit: func(in *OpenTransactionInput) { in.RequestedScope = "artifact/other" }},
		{name: "price_context", edit: func(in *OpenTransactionInput) { in.PriceContext = "quote:9.00-usdc" }},
		{name: "trust_context", edit: func(in *OpenTransactionInput) { in.TrustContext = "trust:0.01" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			ctx := context.Background()
			base := OpenTransactionInput{
				TransactionID:  "deal-conflict-" + tt.name,
				Counterparty:   "did:lango:peer",
				RequestedScope: "artifact/baseline",
				PriceContext:   "quote:1.00-usdc",
				TrustContext:   "trust:0.80",
			}
			_, err := store.OpenKnowledgeExchangeTransaction(ctx, base)
			require.NoError(t, err)

			conflict := base
			tt.edit(&conflict)
			_, err = store.OpenKnowledgeExchangeTransaction(ctx, conflict)
			require.ErrorIs(t, err, ErrInvalidSubmissionInput)
		})
	}
}

func TestApplySettlementProgression_MapsReleaseOutcomeToCanonicalState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-1")

	updated, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "Artifact release approved after review.", "")
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementProgressionReasonCodeApprove, updated.SettlementProgressionReasonCode)
	require.Equal(t, "Artifact release approved after review.", updated.SettlementProgressionReason)
	require.Equal(t, SettlementPending, updated.CanonicalSettlementStatus)

	gotSubmission, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, submission.SubmissionReceiptID, gotSubmission.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, EventSettlementUpdated, events[0].Type)
	require.Equal(t, "settlement_progression", events[0].Source)
	require.Equal(t, string(SettlementProgressionApprovedForSettlement), events[0].Subtype)
	require.Equal(t, "Artifact release approved after review.", events[0].Reason)
}

func TestApplySettlementProgression_AllowsRecoveryFromReviewNeededToApprovedForSettlement(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-1-recovery")

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeRequestRevision, "review", "")
	require.NoError(t, err)

	updated, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "recover", "")
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementPending, updated.CanonicalSettlementStatus)
}

func TestApplySettlementProgression_MapsCanonicalSettlementStatusFromProgressionUpdate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		steps         []SettlementProgressionStatus
		wantCanonical SettlementStatus
	}{
		{
			name:          "approved-for-settlement maps to pending",
			steps:         []SettlementProgressionStatus{SettlementProgressionApprovedForSettlement},
			wantCanonical: SettlementPending,
		},
		{
			name:          "partially-settled maps to partially-settled",
			steps:         []SettlementProgressionStatus{SettlementProgressionApprovedForSettlement, SettlementProgressionPartiallySettled},
			wantCanonical: SettlementPartiallySettled,
		},
		{
			name:          "settled maps to settled",
			steps:         []SettlementProgressionStatus{SettlementProgressionApprovedForSettlement, SettlementProgressionSettled},
			wantCanonical: SettlementSettled,
		},
		{
			name:          "dispute-ready maps to disputed",
			steps:         []SettlementProgressionStatus{SettlementProgressionApprovedForSettlement, SettlementProgressionPartiallySettled, SettlementProgressionDisputeReady},
			wantCanonical: SettlementDisputed,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStore(t)
			ctx := context.Background()
			var err error

			_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-map-"+tc.name)

			var updated TransactionReceipt
			for i, step := range tc.steps {
				reason := "step"
				reasonCode := SettlementProgressionReasonCodeApprove
				if i == len(tc.steps)-1 {
					reason = "final"
				}
				if step == SettlementProgressionReviewNeeded || step == SettlementProgressionDisputeReady {
					reasonCode = SettlementProgressionReasonCodeEscalate
				}
				updated, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, step, reasonCode, reason, "")
				require.NoError(t, err)
			}

			require.Equal(t, tc.steps[len(tc.steps)-1], updated.SettlementProgressionStatus)
			require.Equal(t, tc.wantCanonical, updated.CanonicalSettlementStatus)
		})
	}
}

func TestApplySettlementProgression_RejectsIllegalRewind(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-2")

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "Artifact release approved.", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionPending, SettlementProgressionReasonCodeReject, "rewind", "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestApplySettlementProgression_PreservesStableReasonCodeAndHumanReason(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-escalate-code")

	escalated, err := store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		SettlementProgressionReviewNeeded,
		SettlementProgressionReasonCodeEscalate,
		"manual approval required",
		"",
	)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionReviewNeeded, escalated.SettlementProgressionStatus)
	require.Equal(t, SettlementProgressionReasonCodeEscalate, escalated.SettlementProgressionReasonCode)
	require.NotEqual(t, SettlementProgressionReasonCodeReject, escalated.SettlementProgressionReasonCode)
	require.NotEqual(t, SettlementProgressionReasonCodeRequestRevision, escalated.SettlementProgressionReasonCode)
	require.Equal(t, "manual approval required", escalated.SettlementProgressionReason)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionReasonCodeEscalate, stored.SettlementProgressionReasonCode)
	require.Equal(t, "manual approval required", stored.SettlementProgressionReason)
}

func TestSettlementProgressionHelpers_CoverBoundaryBranches(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateSettlementProgressionTransition("", SettlementProgressionPending))
	require.NoError(t, validateSettlementProgressionTransition(SettlementProgressionReviewNeeded, SettlementProgressionReviewNeeded))
	require.NoError(t, validateSettlementProgressionTransition(SettlementProgressionInProgress, SettlementProgressionReviewNeeded))
	require.NoError(t, validateSettlementProgressionTransition(SettlementProgressionPartiallySettled, SettlementProgressionSettled))
	require.ErrorIs(t, validateSettlementProgressionTransition(SettlementProgressionSettled, SettlementProgressionPending), ErrInvalidSettlementProgressionState)

	require.NoError(t, validateSettlementProgressionReasonCode(SettlementProgressionDisputeReady, SettlementProgressionReasonCodeReject))
	require.NoError(t, validateSettlementProgressionReasonCode(SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeRequestRevision))
	require.NoError(t, validateSettlementProgressionReasonCode(SettlementProgressionInProgress, SettlementProgressionReasonCodeApprove))
	require.NoError(t, validateSettlementProgressionReasonCode(SettlementProgressionPending, ""))
	require.ErrorIs(t, validateSettlementProgressionReasonCode(SettlementProgressionDisputeReady, SettlementProgressionReasonCodeApprove), ErrInvalidSettlementProgressionState)

	require.Equal(t, DisputeLifecycleReEscalated, activeDisputeLifecycleStatus(DisputeLifecycleReEscalated))
	require.Equal(t, DisputeLifecycleHoldActive, activeDisputeLifecycleStatus(""))
	require.Equal(t, DisputeLifecycleStatus(""), nextDisputeLifecycleStatusForProgression(DisputeLifecycleHoldActive, SettlementProgressionDisputeReady, SettlementProgressionSettled))
	require.Equal(t, DisputeLifecycleReEscalated, nextDisputeLifecycleStatusForProgression(DisputeLifecycleHoldActive, SettlementProgressionDisputeReady, SettlementProgressionDisputeReady))
	require.Equal(t, DisputeLifecycleReEscalated, nextDisputeLifecycleStatusForProgression("", SettlementProgressionApprovedForSettlement, SettlementProgressionDisputeReady))
	require.Equal(t, "existing", nextPartialSettlementHint("existing", "   ", SettlementProgressionReviewNeeded))
	require.Empty(t, nextPartialSettlementHint("existing", "", SettlementProgressionSettled))
}

func TestApplySettlementProgression_RequiresCurrentSubmissionReceipt(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-settle-missing-current",
		Counterparty:   "did:lango:peer-missing-current",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "Artifact release approved.", "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestApplySettlementProgression_RejectsInvalidReasonCodeForReviewNeeded(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-invalid-review-needed")

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeApprove, "Artifact release requires review.", "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestApplySettlementProgression_RejectsInvalidReasonCodeForApprovedForSettlement(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-invalid-approved")

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeReject, "Approved after review.", "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestApplySettlementProgression_PersistsPartialHint(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-partial-hint")

	updated, err := store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		SettlementProgressionReviewNeeded,
		SettlementProgressionReasonCodeRequestRevision,
		"Need updated evidence.",
		"settle 40% now, defer the rest",
	)
	require.NoError(t, err)
	require.Equal(t, "settle 40% now, defer the rest", updated.PartialSettlementHint)
}

func TestApplySettlementProgression_AppendsDisputeEventToCurrentSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var err error

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-dispute")

	_, err = store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		SettlementProgressionReviewNeeded,
		SettlementProgressionReasonCodeEscalate,
		"manual review required",
		"",
	)
	require.NoError(t, err)

	updated, err := store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		SettlementProgressionDisputeReady,
		SettlementProgressionReasonCodeEscalate,
		"dispute opened",
		"",
	)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.True(t, updated.DisputeReady)
	require.Equal(t, SettlementDisputed, updated.CanonicalSettlementStatus)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 3)
	require.Equal(t, EventSettlementUpdated, events[0].Type)
	require.Equal(t, EventSettlementUpdated, events[1].Type)
	require.Equal(t, EventDisputed, events[2].Type)
	require.Equal(t, string(SettlementProgressionDisputeReady), events[1].Subtype)
	require.Equal(t, "dispute opened", events[2].Reason)
}

func TestMarkSettlementSettled_ClosesApprovedProgression(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-closeout")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	updated, err := store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		RuntimeReference:     "settlement-tx-123",
	})
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionSettled, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementSettled, updated.CanonicalSettlementStatus)
}

func TestMarkSettlementSettled_AppendsSettlementExecutionTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-closeout-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		RuntimeReference:     "settlement-tx-123",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "settlement_execution", last.Source)
	require.Equal(t, "settled", last.Subtype)
	require.Equal(t, "settlement-tx-123", last.Reason)
}

func TestMarkEscrowReleaseSettled_ClosesApprovedProgression(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-release-closeout")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	updated, err := store.MarkEscrowReleaseSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		RuntimeReference:     "escrow-release-tx-123",
	})
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionSettled, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementSettled, updated.CanonicalSettlementStatus)
}

func TestMarkEscrowReleaseSettled_AppendsEscrowReleaseSuccessEvidence(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-release-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	_, err = store.MarkEscrowReleaseSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		RuntimeReference:     "escrow-release-tx-123",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "escrow_release", last.Source)
	require.Equal(t, "settled", last.Subtype)
	require.Equal(t, "escrow-release-tx-123", last.Reason)
}

func TestRecordSettlementFailure_DoesNotMutateProgression(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-failure")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordSettlementFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, stored.SettlementProgressionStatus)
	require.Equal(t, SettlementPending, stored.CanonicalSettlementStatus)
}

func TestRecordSettlementFailure_AppendsFailureTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-failure-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordSettlementFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
	require.Equal(t, "settlement_execution", last.Source)
	require.Equal(t, "failed", last.Subtype)
	require.Equal(t, "rpc timeout", last.Reason)
}

func TestRecordEscrowReleaseFailure_DoesNotMutateProgression(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-release-failure")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordEscrowReleaseFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, stored.SettlementProgressionStatus)
	require.Equal(t, SettlementPending, stored.CanonicalSettlementStatus)
}

func TestRecordEscrowReleaseFailure_AppendsFailureTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-release-failure-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordEscrowReleaseFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
	require.Equal(t, "escrow_release", last.Source)
	require.Equal(t, "failed", last.Subtype)
	require.Equal(t, "rpc timeout", last.Reason)
}

func TestRecordEscrowRefundSuccess_AppendsRefundTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-refund-success-trail")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow refund test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "refund review", "")
	require.NoError(t, err)

	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "refund-tx-123",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 5)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "escrow_refund", last.Source)
	require.Equal(t, "refunded", last.Subtype)
	require.Equal(t, "refund-tx-123", last.Reason)
}

func TestRecordEscrowRefundSuccess_RejectsWrongState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-refund-success-invalid")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow refund test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "refund-tx-123",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestRecordEscrowRefundFailure_RejectsWrongState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-refund-failure-invalid")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow refund test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "refund failed",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestEscrowExecutionHelpers_CoverValidationAndTransitionBranches(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateEscrowExecutionTransition("", EscrowExecutionStatusPending))
	require.NoError(t, validateEscrowExecutionTransition(EscrowExecutionStatusPending, EscrowExecutionStatusPending))
	require.NoError(t, validateEscrowExecutionTransition(EscrowExecutionStatusCreated, EscrowExecutionStatusFailed))
	require.NoError(t, validateEscrowExecutionTransition(EscrowExecutionStatusFunded, EscrowExecutionStatusFailed))
	require.ErrorIs(t, validateEscrowExecutionTransition(EscrowExecutionStatusFailed, EscrowExecutionStatusCreated), ErrInvalidEscrowExecutionState)

	_, err := escrowExecutionEventTypeForStatus(EscrowExecutionStatus("bogus"))
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionStatus)

	_, err = canonicalizePartialSettlementHint("0")
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	_, err = canonicalizePartialSettlementHint("not-usdc")
	require.Error(t, err)
}

func TestRecordEscrowDisputeHoldSuccess_DoesNotMutateState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-success")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow dispute hold test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute hold review", "")
	require.NoError(t, err)

	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, "escrow-123", updated.EscrowReference)
	require.True(t, updated.DisputeReady)
}

func TestRecordEscrowDisputeHoldSuccess_AppendsHoldTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-success-trail")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow dispute hold test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute hold review", "")
	require.NoError(t, err)

	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 7)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "dispute_hold", last.Source)
	require.Equal(t, "held", last.Subtype)
	require.Equal(t, "hold-123", last.Reason)
}

func TestRecordEscrowDisputeHoldFailure_DoesNotMutateState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-failure")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow dispute hold test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute hold review", "")
	require.NoError(t, err)

	err = store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Reason:               "hold failed",
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, "escrow-123", updated.EscrowReference)
	require.True(t, updated.DisputeReady)
}

func TestRecordEscrowDisputeHoldFailure_AppendsFailureTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-dispute-hold-failure-trail")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow dispute hold test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute hold review", "")
	require.NoError(t, err)

	err = store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Reason:               "hold failed",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 7)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
	require.Equal(t, "dispute_hold", last.Source)
	require.Equal(t, "failed", last.Subtype)
	require.Equal(t, "hold failed", last.Reason)
}

func TestApplyEscrowAdjudication_SetsDecisionWithoutMutatingState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow adjudication test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)

	updated, err := store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRelease,
		Reason:               "fulfilled after review",
	})
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
	require.Equal(t, EscrowAdjudicationRelease, updated.EscrowAdjudication)
	require.Equal(t, DisputeLifecycleHoldActive, updated.DisputeLifecycleStatus)
	require.False(t, updated.DisputeReady)
}

func TestRecordPostAdjudicationDeadLetter_ReEscalatesCanonicalState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-reescalation")
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "re-escalation test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRelease,
		Reason:               "release adjudicated",
	})
	require.NoError(t, err)

	err = store.RecordPostAdjudicationDeadLetter(ctx, PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         3,
		Reason:               "executor timed out repeatedly",
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementDisputed, updated.CanonicalSettlementStatus)
	require.Equal(t, SettlementProgressionReasonCodeEscalate, updated.SettlementProgressionReasonCode)
	require.Equal(t, "post-adjudication execution dead-lettered", updated.SettlementProgressionReason)
	require.True(t, updated.DisputeReady)
	require.Equal(t, DisputeLifecycleReEscalated, updated.DisputeLifecycleStatus)
}

func TestApplyEscrowAdjudication_AppendsDecisionTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-trail")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow adjudication test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRefund,
		Reason:               "refund approved",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 9)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "escrow_adjudication", last.Source)
	require.Equal(t, "refund", last.Subtype)
	require.Equal(t, "refund approved", last.Reason)
}

func TestApplyEscrowAdjudication_RejectsWhenHoldEvidenceMissing(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-no-hold")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow adjudication test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)

	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRelease,
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestRecordEscrowAdjudicationFailure_AppendsFailureTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-escrow-adjudication-failure")

	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "escrow adjudication test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)

	err = store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Reason:               "hold evidence missing",
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Empty(t, updated.EscrowAdjudication)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 7)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
	require.Equal(t, "escrow_adjudication", last.Source)
	require.Equal(t, "failed", last.Subtype)
	require.Equal(t, "hold evidence missing", last.Reason)
}

func TestRecordPostAdjudicationRetryScheduled_AppendsEvidenceWithoutMutatingState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-retry")
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "retry test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRelease,
		Reason:               "release adjudicated",
	})
	require.NoError(t, err)

	nextRetryAt := time.Now().Add(2 * time.Minute).UTC().Truncate(time.Second)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          nextRetryAt,
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowAdjudicationRelease, updated.EscrowAdjudication)
	require.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	last := events[len(events)-1]
	require.Equal(t, "post_adjudication_retry", last.Source)
	require.Equal(t, "retry-scheduled", last.Subtype)
	require.Equal(t, EventSettlementUpdated, last.Type)
}

func TestRecordPostAdjudicationDeadLetter_AppendsTerminalFailureAndReEscalatesState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-post-adjudication-dead-letter")
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "retry test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRefund,
		Reason:               "refund adjudicated",
	})
	require.NoError(t, err)

	err = store.RecordPostAdjudicationDeadLetter(ctx, PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRefund,
		AttemptCount:         4,
		Reason:               "worker failed repeatedly",
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowAdjudicationRefund, updated.EscrowAdjudication)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementDisputed, updated.CanonicalSettlementStatus)
	require.Equal(t, DisputeLifecycleReEscalated, updated.DisputeLifecycleStatus)
	require.True(t, updated.DisputeReady)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	last := events[len(events)-1]
	require.Equal(t, "post_adjudication_retry", last.Source)
	require.Equal(t, "dead-lettered", last.Subtype)
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
}

func TestRecordManualRetryRequested_AppendsEvidenceWithoutDroppingReEscalatedState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-manual-retry-requested")
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "0.50",
		Reason:    "manual retry test",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "0.50"},
		},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-123", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-123",
		Outcome:              EscrowAdjudicationRefund,
		Reason:               "refund adjudicated",
	})
	require.NoError(t, err)
	err = store.RecordPostAdjudicationDeadLetter(ctx, PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRefund,
		AttemptCount:         4,
		Reason:               "worker failed repeatedly",
	})
	require.NoError(t, err)

	err = store.RecordManualRetryRequested(ctx, ManualRetryRequestedRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRefund,
	})
	require.NoError(t, err)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, EscrowAdjudicationRefund, updated.EscrowAdjudication)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementDisputed, updated.CanonicalSettlementStatus)
	require.Equal(t, DisputeLifecycleReEscalated, updated.DisputeLifecycleStatus)
	require.True(t, updated.DisputeReady)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	last := events[len(events)-1]
	require.Equal(t, "post_adjudication_retry", last.Source)
	require.Equal(t, "manual-retry-requested", last.Subtype)
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Contains(t, last.Reason, "manual_replay_at=")
}

func TestRecordSettlementFailure_RejectsFailureAfterSettlementCloseout(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-settle-failure-after-closeout")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		RuntimeReference:     "settlement-tx-123",
	})
	require.NoError(t, err)

	err = store.RecordSettlementFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
}

func TestMarkPartialSettlementSettled_ClosesApprovedProgressionToPartiallySettled(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-closeout")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	updated, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.60",
		RuntimeReference:     "partial-tx-123",
	})
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionPartiallySettled, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementPartiallySettled, updated.CanonicalSettlementStatus)
}

func TestApplySettlementProgression_ReEscalationFromPartiallySettledPreservesPartialHint(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-settlement-reescalation")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)
	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.60",
		RuntimeReference:     "partial-123",
	})
	require.NoError(t, err)

	updated, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "renewed disagreement", "")
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionDisputeReady, updated.SettlementProgressionStatus)
	require.Equal(t, SettlementDisputed, updated.CanonicalSettlementStatus)
	require.Equal(t, "settle:0.60-usdc", updated.PartialSettlementHint)
	require.True(t, updated.DisputeReady)
}

func TestMarkPartialSettlementSettled_CanonicalizesRemainingHint(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-remaining-hint")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	updated, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.6000",
		RuntimeReference:     "partial-tx-123",
	})
	require.NoError(t, err)
	require.Equal(t, "settle:0.60-usdc", updated.PartialSettlementHint)
}

func TestMarkPartialSettlementSettled_AppendsSuccessTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-success-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.60",
		RuntimeReference:     "partial-tx-123",
	})
	require.NoError(t, err)

	err = store.RecordPartialSettlementSuccess(ctx, PartialSettlementExecutionEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "partial-tx-123",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementUpdated, last.Type)
	require.Equal(t, "partial_settlement_execution", last.Source)
	require.Equal(t, "partially-settled", last.Subtype)
	require.Equal(t, "partial-tx-123", last.Reason)
}

func TestRecordPartialSettlementFailure_DoesNotMutateProgression(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-failure")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordPartialSettlementFailure(ctx, PartialSettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.60",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionApprovedForSettlement, stored.SettlementProgressionStatus)
	require.Equal(t, SettlementPending, stored.CanonicalSettlementStatus)
}

func TestRecordPartialSettlementFailure_AppendsFailureTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	submission, tx := createSubmittedTransaction(t, store, ctx, "deal-partial-failure-trail")

	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	err = store.RecordPartialSettlementFailure(ctx, PartialSettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.40",
		RemainingAmount:      "0.60",
		Reason:               "rpc timeout",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	last := events[len(events)-1]
	require.Equal(t, EventSettlementExecutionFailed, last.Type)
	require.Equal(t, "partial_settlement_execution", last.Source)
	require.Equal(t, "failed", last.Subtype)
	require.Equal(t, "rpc timeout", last.Reason)
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

func TestListSubmissionReceipts_ReturnsAllSubmissionsAcrossTransactionHistory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-list-submissions",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-list-submissions",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)

	got, err := store.ListSubmissionReceipts(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.ElementsMatch(t, []SubmissionReceipt{first, second}, got)
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

func TestAppendPaymentExecutionEvents_RecordAuthorizedAndDeniedTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	sub, _ := createSubmittedTransaction(t, store, ctx, "tx-payment-events")

	require.NoError(t, store.AppendPaymentExecutionAuthorized(ctx, sub.SubmissionReceiptID))
	require.NoError(t, store.AppendPaymentExecutionDenied(ctx, sub.SubmissionReceiptID, "policy denied"))

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, ReceiptEvent{
		SubmissionReceiptID: sub.SubmissionReceiptID,
		Source:              "payment_execution",
		Subtype:             "authorized",
		Type:                EventPaymentExecutionAuthorized,
	}, events[0])
	assert.Equal(t, ReceiptEvent{
		SubmissionReceiptID: sub.SubmissionReceiptID,
		Source:              "payment_execution",
		Subtype:             "denied",
		Reason:              "policy denied",
		Type:                EventPaymentExecutionDenied,
	}, events[1])
}

func TestAppendPaymentExecutionEvents_RejectMissingSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AppendPaymentExecutionAuthorized(ctx, "missing")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	err = store.AppendPaymentExecutionDenied(ctx, "missing", "no receipt")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestListTransactionReceipts_ReturnsClonedEscrowExecutionInputs(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	sub, tx := createSubmittedTransaction(t, store, ctx, "tx-list-clone")
	input := EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "3.00",
		Reason:    "escrow",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "1.00"},
			{Description: "final", Amount: "2.00"},
		},
	}
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, input)
	require.NoError(t, err)

	listed, err := store.ListTransactionReceipts(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.NotNil(t, listed[0].EscrowExecutionInput)
	listed[0].EscrowExecutionInput.Milestones[0].Description = "mutated"

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.NotNil(t, stored.EscrowExecutionInput)
	assert.Equal(t, "draft", stored.EscrowExecutionInput.Milestones[0].Description)
}

func TestMarkSettlementPartiallySettled_AliasAndSuccessTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	sub, tx := createSubmittedTransaction(t, store, ctx, "tx-partial-alias")
	approved, err := store.ApplySettlementProgression(
		ctx,
		tx.TransactionReceiptID,
		SettlementProgressionApprovedForSettlement,
		SettlementProgressionReasonCodeApprove,
		"approved",
		"",
	)
	require.NoError(t, err)

	partial, err := store.MarkSettlementPartiallySettled(ctx, SettlementPartialCloseoutRequest{
		TransactionReceiptID: approved.TransactionReceiptID,
		SubmissionReceiptID:  sub.SubmissionReceiptID,
		ExecutedAmount:       "1.00",
		RemainingAmount:      "2.50",
		RuntimeReference:     "partial-runtime",
	})
	require.NoError(t, err)
	assert.Equal(t, SettlementProgressionPartiallySettled, partial.SettlementProgressionStatus)
	assert.Equal(t, SettlementPartiallySettled, partial.CanonicalSettlementStatus)
	assert.Equal(t, "settle:2.50-usdc", partial.PartialSettlementHint)

	err = store.RecordSettlementPartialSuccess(ctx, SettlementPartialExecutionEvidenceRequest{
		TransactionReceiptID: partial.TransactionReceiptID,
		SubmissionReceiptID:  sub.SubmissionReceiptID,
		RuntimeReference:     "partial-runtime",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.NotEmpty(t, events)
	last := events[len(events)-1]
	assert.Equal(t, "partial_settlement_execution", last.Source)
	assert.Equal(t, "partially-settled", last.Subtype)
	assert.Equal(t, "partial-runtime", last.Reason)
	assert.Equal(t, EventSettlementUpdated, last.Type)
}

func TestCreateSubmissionReceipt_RejectsEachMissingRequiredField(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	base := CreateSubmissionInput{
		TransactionID:       "tx-validation",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	}
	tests := []struct {
		name   string
		mutate func(*CreateSubmissionInput)
		want   string
	}{
		{
			name: "missing transaction id",
			mutate: func(in *CreateSubmissionInput) {
				in.TransactionID = ""
			},
			want: "transaction_id is required",
		},
		{
			name: "missing artifact label",
			mutate: func(in *CreateSubmissionInput) {
				in.ArtifactLabel = ""
			},
			want: "artifact_label is required",
		},
		{
			name: "missing payload hash",
			mutate: func(in *CreateSubmissionInput) {
				in.PayloadHash = ""
			},
			want: "payload_hash is required",
		},
		{
			name: "missing source lineage digest",
			mutate: func(in *CreateSubmissionInput) {
				in.SourceLineageDigest = ""
			},
			want: "source_lineage_digest is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base
			tt.mutate(&input)
			_, _, err := store.CreateSubmissionReceipt(ctx, input)
			require.ErrorIs(t, err, ErrInvalidSubmissionInput)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestApplyKnowledgeExchangeRuntimeProgression_CoversLegalTerminalBranches(t *testing.T) {
	tests := []struct {
		name     string
		terminal KnowledgeExchangeRuntimeStatus
	}{
		{name: "release approved", terminal: RuntimeStatusReleaseApproved},
		{name: "escalated", terminal: RuntimeStatusEscalated},
		{name: "dispute ready", terminal: RuntimeStatusDisputeReady},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			ctx := context.Background()
			transactionID := "tx-runtime-" + string(tt.terminal)
			tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
				TransactionID:  transactionID,
				Counterparty:   "did:lango:peer",
				RequestedScope: "artifact/runtime",
			})
			require.NoError(t, err)
			sub, _ := createSubmittedTransaction(t, store, ctx, transactionID)

			for _, next := range []KnowledgeExchangeRuntimeStatus{
				RuntimeStatusExportabilityAdvisory,
				RuntimeStatusPaymentApproved,
				RuntimeStatusPaymentAuthorized,
				RuntimeStatusWorkStarted,
				RuntimeStatusSubmissionReceived,
				tt.terminal,
			} {
				submissionID := ""
				if next == RuntimeStatusSubmissionReceived {
					submissionID = sub.SubmissionReceiptID
				}
				tx, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, next, submissionID)
				require.NoError(t, err)
			}

			assert.Equal(t, tt.terminal, tx.KnowledgeExchangeRuntimeStatus)
			assert.Equal(t, sub.SubmissionReceiptID, tx.CurrentSubmissionReceiptID)
		})
	}
}
