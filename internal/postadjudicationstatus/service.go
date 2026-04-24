package postadjudicationstatus

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/langoai/lango/internal/receipts"
)

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) ListCurrentDeadLetters(ctx context.Context) ([]DeadLetterBacklogEntry, error) {
	page, err := s.ListCurrentDeadLettersPage(ctx, DeadLetterListOptions{})
	if err != nil {
		return nil, err
	}
	return append([]DeadLetterBacklogEntry(nil), page.Items...), nil
}

func (s *Service) ListCurrentDeadLettersPage(ctx context.Context, opts DeadLetterListOptions) (DeadLetterListPage, error) {
	if s == nil || s.store == nil {
		return DeadLetterListPage{}, fmt.Errorf("receipt store is required")
	}

	transactions, err := s.store.ListTransactionReceipts(ctx)
	if err != nil {
		return DeadLetterListPage{}, err
	}

	entries := make([]DeadLetterBacklogEntry, 0, len(transactions))
	for _, transaction := range transactions {
		entry, ok, err := s.deadLetterEntryForTransaction(ctx, transaction)
		if err != nil {
			return DeadLetterListPage{}, err
		}
		if !ok {
			continue
		}
		if !matchesDeadLetterFilters(entry, opts) {
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

	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > len(entries) {
		offset = len(entries)
	}

	limit := opts.Limit
	if limit < 0 {
		limit = 0
	}

	end := len(entries)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	pageItems := entries[offset:end]

	return DeadLetterListPage{
		Items:  append([]DeadLetterBacklogEntry(nil), pageItems...),
		Total:  len(entries),
		Count:  len(pageItems),
		Offset: offset,
		Limit:  limit,
	}, nil
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
	isDeadLettered, canRetry, adjudication := detailNavigationHints(transaction, summary)
	return TransactionStatus{
		CanonicalSnapshot: CanonicalSnapshot{
			TransactionReceipt: transaction,
			SubmissionReceipt:  submission,
			SubmissionEvents:   append([]receipts.ReceiptEvent(nil), events...),
		},
		RetryDeadLetterSummary: summary,
		IsDeadLettered:         isDeadLettered,
		CanRetry:               canRetry,
		Adjudication:           adjudication,
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
	isDeadLettered, canRetry, adjudication := detailNavigationHints(transaction, summary)

	return DeadLetterBacklogEntry{
		TransactionReceiptID:    transaction.TransactionReceiptID,
		SubmissionReceiptID:     submissionReceiptID,
		Adjudication:            adjudication,
		IsDeadLettered:          isDeadLettered,
		CanRetry:                canRetry,
		LatestDeadLetterReason:  summary.LatestDeadLetterReason,
		LatestDeadLetteredAt:    summary.LatestDeadLetteredAt,
		LatestManualReplayActor: summary.LatestManualReplayActor,
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
		if event.Subtype == "manual-retry-requested" && parsed.LatestManualReplayActor != "" {
			summary.LatestManualReplayActor = parsed.LatestManualReplayActor
		}
		summary.LatestStatusSubtype = event.Subtype
		if event.Subtype == "dead-lettered" {
			summary.HasDeadLetter = true
			if parsed.LatestDeadLetterReason != "" {
				summary.LatestDeadLetterReason = parsed.LatestDeadLetterReason
			}
			if parsed.LatestDeadLetteredAt != "" {
				summary.LatestDeadLetteredAt = parsed.LatestDeadLetteredAt
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
		case "actor":
			summary.LatestManualReplayActor = value
		case "dead_lettered_at":
			if deadLetteredAt, err := time.Parse(time.RFC3339, value); err == nil {
				summary.LatestDeadLetteredAt = deadLetteredAt.UTC().Format(time.RFC3339)
			}
		}
	}

	return summary
}

func matchesDeadLetterFilters(entry DeadLetterBacklogEntry, opts DeadLetterListOptions) bool {
	if adjudication := strings.TrimSpace(opts.Adjudication); adjudication != "" && !strings.EqualFold(entry.Adjudication, adjudication) {
		return false
	}
	if opts.RetryAttemptMin > 0 && entry.LatestRetryAttempt < opts.RetryAttemptMin {
		return false
	}
	if opts.RetryAttemptMax > 0 && entry.LatestRetryAttempt > opts.RetryAttemptMax {
		return false
	}
	if query := strings.TrimSpace(opts.Query); query != "" {
		needle := strings.ToLower(query)
		haystack := strings.ToLower(entry.TransactionReceiptID + " " + entry.SubmissionReceiptID)
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	if actor := strings.TrimSpace(opts.ManualReplayActor); actor != "" && !strings.EqualFold(entry.LatestManualReplayActor, actor) {
		return false
	}
	if after := strings.TrimSpace(opts.DeadLetteredAfter); after != "" {
		afterTime, ok := parseRFC3339(after)
		if !ok {
			return false
		}
		entryTime, ok := parseRFC3339(entry.LatestDeadLetteredAt)
		if !ok || entryTime.Before(afterTime) {
			return false
		}
	}
	if before := strings.TrimSpace(opts.DeadLetteredBefore); before != "" {
		beforeTime, ok := parseRFC3339(before)
		if !ok {
			return false
		}
		entryTime, ok := parseRFC3339(entry.LatestDeadLetteredAt)
		if !ok || entryTime.After(beforeTime) {
			return false
		}
	}
	if reasonQuery := strings.TrimSpace(opts.DeadLetterReasonQuery); reasonQuery != "" {
		if !strings.Contains(strings.ToLower(entry.LatestDeadLetterReason), strings.ToLower(reasonQuery)) {
			return false
		}
	}
	if dispatchReference := strings.TrimSpace(opts.LatestDispatchReference); dispatchReference != "" && entry.LatestDispatchReference != dispatchReference {
		return false
	}
	return true
}

func detailNavigationHints(transaction receipts.TransactionReceipt, summary RetryDeadLetterSummary) (bool, bool, string) {
	adjudication := string(transaction.EscrowAdjudication)
	isDeadLettered := summary.HasDeadLetter
	canRetry := isDeadLettered && adjudication != "" && strings.TrimSpace(transaction.CurrentSubmissionReceiptID) != ""
	return isDeadLettered, canRetry, adjudication
}

func parseRFC3339(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}
