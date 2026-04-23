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
	store      receiptStore
	dispatcher dispatcher
	policy     ReplayPolicy
}

func NewService(store receiptStore, dispatcher dispatcher, policy ...ReplayPolicy) *Service {
	resolved := ReplayPolicy{}
	if len(policy) > 0 {
		resolved = policy[0]
	}
	return &Service{
		store:      store,
		dispatcher: dispatcher,
		policy:     resolved,
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

	if !hasDeadLetterEvidence(events) {
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

	canonical.SubmissionEvents = append(canonical.SubmissionEvents, receipts.ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "post_adjudication_retry",
		Subtype:             "manual-retry-requested",
		Reason:              "manual retry requested",
		Type:                receipts.EventSettlementUpdated,
	})

	dispatchReceipt, err := s.dispatcher.Dispatch(ctx, BackgroundDispatchRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      strings.TrimSpace(transaction.EscrowReference),
		Outcome:              transaction.EscrowAdjudication,
		Prompt:               buildBackgroundDispatchPrompt(transaction, submission),
	})
	if err != nil {
		return Result{CanonicalAdjudication: canonical}, fmt.Errorf("dispatch background post-adjudication: %w", err)
	}

	return Result{
		CanonicalAdjudication:     canonical,
		BackgroundDispatchReceipt: &dispatchReceipt,
	}, nil
}

func buildBackgroundDispatchPrompt(transaction receipts.TransactionReceipt, submission receipts.SubmissionReceipt) string {
	toolName := "release_escrow_settlement"
	switch transaction.EscrowAdjudication {
	case receipts.EscrowAdjudicationRefund:
		toolName = "refund_escrow_settlement"
	}

	return fmt.Sprintf(
		"Execute the adjudicated escrow %s branch for transaction_receipt_id=%s.\nUse %s to perform the branch as a background follow-up.\nThe canonical adjudication is already recorded for submission_receipt_id=%s and escrow_reference=%s.\nDo not re-adjudicate.",
		transaction.EscrowAdjudication,
		transaction.TransactionReceiptID,
		toolName,
		submission.SubmissionReceiptID,
		strings.TrimSpace(transaction.EscrowReference),
	)
}

func hasDeadLetterEvidence(events []receipts.ReceiptEvent) bool {
	for _, event := range events {
		if event.Source == "post_adjudication_retry" && event.Subtype == "dead-lettered" {
			return true
		}
	}
	return false
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
