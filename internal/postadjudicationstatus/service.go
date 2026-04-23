package postadjudicationstatus

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/receipts"
)

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) ListCurrentDeadLetters(ctx context.Context) ([]DeadLetterBacklogEntry, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("receipt store is required")
	}

	transactions, err := s.store.ListTransactionReceipts(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]DeadLetterBacklogEntry, 0, len(transactions))
	for _, transaction := range transactions {
		entry, ok, err := s.deadLetterEntryForTransaction(ctx, transaction)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].LatestRetryAttempt != entries[j].LatestRetryAttempt {
			return entries[i].LatestRetryAttempt > entries[j].LatestRetryAttempt
		}
		return entries[i].TransactionReceiptID < entries[j].TransactionReceiptID
	})

	return entries, nil
}

func (s *Service) GetTransactionStatus(ctx context.Context, transactionReceiptID string) (TransactionStatus, error) {
	if s == nil || s.store == nil {
		return TransactionStatus{}, fmt.Errorf("receipt store is required")
	}

	transactionReceiptID = strings.TrimSpace(transactionReceiptID)
	if transactionReceiptID == "" {
		return TransactionStatus{}, ErrTransactionReceiptNotFound
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrTransactionReceiptNotFound) {
			return TransactionStatus{}, ErrTransactionReceiptNotFound
		}
		return TransactionStatus{}, err
	}

	submission, events, err := s.currentCanonicalSnapshot(ctx, transaction)
	if err != nil {
		return TransactionStatus{}, err
	}

	summary := summarizeEvents(events)
	return TransactionStatus{
		CanonicalSnapshot: CanonicalSnapshot{
			TransactionReceipt: transaction,
			SubmissionReceipt:  submission,
			SubmissionEvents:   append([]receipts.ReceiptEvent(nil), events...),
		},
		RetryDeadLetterSummary: summary,
	}, nil
}

func (s *Service) deadLetterEntryForTransaction(ctx context.Context, transaction receipts.TransactionReceipt) (DeadLetterBacklogEntry, bool, error) {
	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return DeadLetterBacklogEntry{}, false, nil
	}

	submission, events, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return DeadLetterBacklogEntry{}, false, nil
		}
		return DeadLetterBacklogEntry{}, false, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return DeadLetterBacklogEntry{}, false, nil
	}

	summary := summarizeEvents(events)
	if !summary.HasDeadLetter {
		return DeadLetterBacklogEntry{}, false, nil
	}

	return DeadLetterBacklogEntry{
		TransactionReceiptID:    transaction.TransactionReceiptID,
		SubmissionReceiptID:     submissionReceiptID,
		Adjudication:            string(transaction.EscrowAdjudication),
		LatestDeadLetterReason:  summary.LatestDeadLetterReason,
		LatestRetryAttempt:      summary.LatestRetryAttempt,
		LatestDispatchReference: summary.LatestDispatchReference,
	}, true, nil
}

func (s *Service) currentCanonicalSnapshot(ctx context.Context, transaction receipts.TransactionReceipt) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error) {
	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return receipts.SubmissionReceipt{}, nil, ErrCurrentSubmissionMissing
	}

	submission, events, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return receipts.SubmissionReceipt{}, nil, ErrCurrentSubmissionMissing
		}
		return receipts.SubmissionReceipt{}, nil, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return receipts.SubmissionReceipt{}, nil, ErrCurrentSubmissionMissing
	}

	return submission, events, nil
}

type eventSummary = RetryDeadLetterSummary

func summarizeEvents(events []receipts.ReceiptEvent) eventSummary {
	var summary eventSummary
	for _, event := range events {
		if event.Source != "post_adjudication_retry" {
			continue
		}

		parsed := parseEventSummary(event)
		if parsed.LatestRetryAttempt > 0 {
			summary.LatestRetryAttempt = parsed.LatestRetryAttempt
		}
		if parsed.LatestDispatchReference != "" {
			summary.LatestDispatchReference = parsed.LatestDispatchReference
		}
		summary.LatestStatusSubtype = event.Subtype
		if event.Subtype == "dead-lettered" {
			summary.HasDeadLetter = true
			if parsed.LatestDeadLetterReason != "" {
				summary.LatestDeadLetterReason = parsed.LatestDeadLetterReason
			}
			continue
		}
		summary.HasDeadLetter = false
	}

	return summary
}

func parseEventSummary(event receipts.ReceiptEvent) eventSummary {
	summary := eventSummary{}
	reason := strings.TrimSpace(event.Reason)
	if reason == "" {
		return summary
	}

	if idx := strings.Index(reason, "reason="); idx >= 0 {
		summary.LatestDeadLetterReason = strings.TrimSpace(reason[idx+len("reason="):])
		reason = strings.TrimSpace(reason[:idx])
	} else {
		summary.LatestDeadLetterReason = reason
	}

	for _, field := range strings.Fields(reason) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		switch key {
		case "attempt":
			if attempt, err := strconv.Atoi(value); err == nil {
				summary.LatestRetryAttempt = attempt
			}
		case "dispatch_reference":
			summary.LatestDispatchReference = value
		}
	}

	return summary
}
