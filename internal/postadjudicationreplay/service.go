package postadjudicationreplay

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/receipts"
)

type Service struct {
	store          receiptStore
	dispatcher     dispatcher
	policy         ReplayPolicy
	recoveryPolicy RecoveryPolicy
}

func NewService(store receiptStore, dispatcher dispatcher, policy ...ReplayPolicy) *Service {
	resolved := ReplayPolicy{}
	if len(policy) > 0 {
		resolved = policy[0]
	}
	return &Service{
		store:          store,
		dispatcher:     dispatcher,
		policy:         resolved,
		recoveryPolicy: DefaultRecoveryPolicy(),
	}
}

func (s *Service) Replay(ctx context.Context, req Request) (Result, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return Result{}, fmt.Errorf("transaction_receipt_id is required")
	}
	if s == nil || s.store == nil {
		return Result{}, fmt.Errorf("receipt store is required")
	}
	if s.dispatcher == nil {
		return Result{}, fmt.Errorf("dispatcher is required")
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrTransactionReceiptNotFound) {
			return Result{}, ErrTransactionReceiptNotFound
		}
		return Result{}, err
	}

	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return Result{}, ErrCurrentSubmissionMissing
	}

	submission, events, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return Result{}, ErrCurrentSubmissionMissing
		}
		return Result{}, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return Result{}, ErrCurrentSubmissionMissing
	}

	if !s.recoveryPolicy.HasDeadLetterEvidence(events) {
		return Result{}, ErrDeadLetterEvidenceMissing
	}
	if transaction.EscrowAdjudication != receipts.EscrowAdjudicationRelease &&
		transaction.EscrowAdjudication != receipts.EscrowAdjudicationRefund {
		return Result{}, ErrCanonicalAdjudicationMissing
	}

	actor := strings.TrimSpace(ctxkeys.PrincipalFromContext(ctx))
	if actor == "" {
		return Result{}, ErrActorUnresolved
	}
	if !replayAllowedForOutcome(actor, transaction.EscrowAdjudication, s.policy) {
		return Result{}, ErrReplayNotAllowed
	}

	canonical := CanonicalAdjudicationSnapshot{
		TransactionReceipt: transaction,
		SubmissionReceipt:  submission,
		SubmissionEvents:   append([]receipts.ReceiptEvent(nil), events...),
	}

	if err := s.store.RecordManualRetryRequested(ctx, receipts.ManualRetryRequestedRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		Outcome:              transaction.EscrowAdjudication,
		Reason:               "manual retry requested",
	}); err != nil {
		return Result{CanonicalAdjudication: canonical}, fmt.Errorf("record manual retry requested: %w", err)
	}

	_, refreshedEvents, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err == nil {
		canonical.SubmissionEvents = append([]receipts.ReceiptEvent(nil), refreshedEvents...)
	}

	dispatchReceipt, err := s.dispatcher.Dispatch(ctx, BackgroundDispatchRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      strings.TrimSpace(transaction.EscrowReference),
		Outcome:              transaction.EscrowAdjudication,
		Prompt: BuildBackgroundDispatchPrompt(
			transaction.EscrowAdjudication,
			transaction.TransactionReceiptID,
			submission.SubmissionReceiptID,
			transaction.EscrowReference,
		),
	})
	if err != nil {
		return Result{CanonicalAdjudication: canonical}, fmt.Errorf("dispatch background post-adjudication: %w", err)
	}

	return Result{
		CanonicalAdjudication:     canonical,
		BackgroundDispatchReceipt: &dispatchReceipt,
	}, nil
}

func replayAllowedForOutcome(actor string, outcome receipts.EscrowAdjudicationDecision, policy ReplayPolicy) bool {
	if !containsActor(policy.AllowedActors, actor) {
		return false
	}

	switch outcome {
	case receipts.EscrowAdjudicationRelease:
		return containsActor(policy.ReleaseAllowedActors, actor)
	case receipts.EscrowAdjudicationRefund:
		return containsActor(policy.RefundAllowedActors, actor)
	default:
		return false
	}
}

func containsActor(actors []string, actor string) bool {
	for _, candidate := range actors {
		if candidate == actor {
			return true
		}
	}
	return false
}
