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

type transactionGlobalAggregation struct {
	RetryDeadLetterSummary
	SubmissionBreakdown []SubmissionBreakdownItem
}

type submissionReceiptLister interface {
	ListSubmissionReceipts(context.Context) ([]receipts.SubmissionReceipt, error)
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

	sortDeadLetterEntries(entries, opts.SortBy)

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
	transactionGlobal, err := s.transactionGlobalRetryAggregation(ctx, transaction)
	if err != nil {
		return DeadLetterBacklogEntry{}, false, err
	}
	isDeadLettered, canRetry, adjudication := detailNavigationHints(transaction, summary)

	return DeadLetterBacklogEntry{
		TransactionReceiptID:              transaction.TransactionReceiptID,
		SubmissionReceiptID:               submissionReceiptID,
		SubmissionBreakdown:               append([]SubmissionBreakdownItem(nil), transactionGlobal.SubmissionBreakdown...),
		Adjudication:                      adjudication,
		IsDeadLettered:                    isDeadLettered,
		CanRetry:                          canRetry,
		LatestDeadLetterReason:            summary.LatestDeadLetterReason,
		LatestDeadLetteredAt:              summary.LatestDeadLetteredAt,
		LatestManualReplayActor:           summary.LatestManualReplayActor,
		LatestManualReplayAt:              summary.LatestManualReplayAt,
		LatestStatusSubtype:               summary.LatestStatusSubtype,
		LatestStatusSubtypeFamily:         summary.LatestStatusSubtypeFamily,
		DominantFamily:                    summary.DominantFamily,
		AnyMatchFamilies:                  append([]string(nil), summary.AnyMatchFamilies...),
		ManualRetryCount:                  summary.ManualRetryCount,
		TotalRetryCount:                   summary.TotalRetryCount,
		TransactionGlobalTotalRetryCount:  transactionGlobal.TotalRetryCount,
		TransactionGlobalAnyMatchFamilies: append([]string(nil), transactionGlobal.AnyMatchFamilies...),
		TransactionGlobalDominantFamily:   transactionGlobal.TransactionGlobalDominantFamily,
		LatestRetryAttempt:                summary.LatestRetryAttempt,
		LatestDispatchReference:           summary.LatestDispatchReference,
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

func (s *Service) transactionGlobalRetryAggregation(ctx context.Context, transaction receipts.TransactionReceipt) (transactionGlobalAggregation, error) {
	submissions, err := s.submissionReceiptsForTransaction(ctx, transaction)
	if err != nil {
		return transactionGlobalAggregation{}, err
	}
	sort.SliceStable(submissions, func(i, j int) bool {
		leftIsCurrent := submissions[i].SubmissionReceiptID == transaction.CurrentSubmissionReceiptID
		rightIsCurrent := submissions[j].SubmissionReceiptID == transaction.CurrentSubmissionReceiptID
		if leftIsCurrent != rightIsCurrent {
			return !leftIsCurrent && rightIsCurrent
		}
		return submissions[i].SubmissionReceiptID < submissions[j].SubmissionReceiptID
	})

	familySet := make(map[string]struct{})
	familyCounts := make(map[string]int)
	totalRetryCount := 0
	latestRelevantFamily := ""
	breakdown := make([]SubmissionBreakdownItem, 0, len(submissions))
	for _, submission := range submissions {
		_, events, err := s.store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
		if err != nil {
			if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
				continue
			}
			return transactionGlobalAggregation{}, err
		}

		summary := summarizeEvents(events)
		totalRetryCount += summary.TotalRetryCount
		for _, family := range summary.AnyMatchFamilies {
			familySet[family] = struct{}{}
		}
		for _, event := range events {
			if event.Source != "post_adjudication_retry" {
				continue
			}
			family := subtypeFamily(event.Subtype)
			if family == "" {
				continue
			}
			familyCounts[family]++
			latestRelevantFamily = family
		}
		breakdown = append(breakdown, SubmissionBreakdownItem{
			SubmissionReceiptID: submission.SubmissionReceiptID,
			RetryCount:          summary.TotalRetryCount,
			AnyMatchFamilies:    append([]string(nil), summary.AnyMatchFamilies...),
		})
	}

	return transactionGlobalAggregation{
		RetryDeadLetterSummary: RetryDeadLetterSummary{
			TotalRetryCount:                 totalRetryCount,
			AnyMatchFamilies:                anyMatchFamilies(familySet),
			TransactionGlobalDominantFamily: dominantFamily(familyCounts, latestRelevantFamily),
		},
		SubmissionBreakdown: breakdown,
	}, nil
}

func (s *Service) submissionReceiptsForTransaction(ctx context.Context, transaction receipts.TransactionReceipt) ([]receipts.SubmissionReceipt, error) {
	if lister, ok := s.store.(submissionReceiptLister); ok {
		submissions, err := lister.ListSubmissionReceipts(ctx)
		if err != nil {
			return nil, err
		}

		filtered := make([]receipts.SubmissionReceipt, 0, len(submissions))
		for _, submission := range submissions {
			if submission.TransactionReceiptID != transaction.TransactionReceiptID {
				continue
			}
			filtered = append(filtered, submission)
		}
		return filtered, nil
	}

	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return nil, nil
	}

	submission, _, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return nil, nil
	}

	return []receipts.SubmissionReceipt{submission}, nil
}

type eventSummary = RetryDeadLetterSummary

func summarizeEvents(events []receipts.ReceiptEvent) eventSummary {
	var summary eventSummary
	familySet := make(map[string]struct{})
	familyCounts := make(map[string]int)
	latestRelevantFamily := ""
	for _, event := range events {
		if event.Source != "post_adjudication_retry" {
			continue
		}
		family := subtypeFamily(event.Subtype)
		if family != "" {
			summary.TotalRetryCount++
			summary.LatestStatusSubtypeFamily = family
			familySet[family] = struct{}{}
			familyCounts[family]++
			latestRelevantFamily = family
		}

		parsed := parseEventSummary(event)
		if parsed.LatestRetryAttempt > 0 {
			summary.LatestRetryAttempt = parsed.LatestRetryAttempt
		}
		if parsed.LatestDispatchReference != "" {
			summary.LatestDispatchReference = parsed.LatestDispatchReference
		}
		if event.Subtype == "manual-retry-requested" {
			summary.ManualRetryCount++
			if parsed.LatestManualReplayActor != "" {
				summary.LatestManualReplayActor = parsed.LatestManualReplayActor
			}
			if parsed.LatestManualReplayAt != "" {
				summary.LatestManualReplayAt = parsed.LatestManualReplayAt
			}
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
	summary.DominantFamily = dominantFamily(familyCounts, latestRelevantFamily)
	summary.AnyMatchFamilies = anyMatchFamilies(familySet)

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
		case "manual_replay_at":
			if manualReplayAt, err := time.Parse(time.RFC3339, value); err == nil {
				summary.LatestManualReplayAt = manualReplayAt.UTC().Format(time.RFC3339)
			}
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
	if latestStatusSubtype := strings.TrimSpace(opts.LatestStatusSubtype); latestStatusSubtype != "" && entry.LatestStatusSubtype != latestStatusSubtype {
		return false
	}
	if latestStatusSubtypeFamily := strings.TrimSpace(opts.LatestStatusSubtypeFamily); latestStatusSubtypeFamily != "" && entry.LatestStatusSubtypeFamily != latestStatusSubtypeFamily {
		return false
	}
	if opts.ManualRetryCountMin > 0 && entry.ManualRetryCount < opts.ManualRetryCountMin {
		return false
	}
	if opts.ManualRetryCountMax > 0 && entry.ManualRetryCount > opts.ManualRetryCountMax {
		return false
	}
	if opts.TotalRetryCountMin > 0 && entry.TotalRetryCount < opts.TotalRetryCountMin {
		return false
	}
	if opts.TotalRetryCountMax > 0 && entry.TotalRetryCount > opts.TotalRetryCountMax {
		return false
	}
	if opts.TransactionGlobalTotalRetryCountMin > 0 && entry.TransactionGlobalTotalRetryCount < opts.TransactionGlobalTotalRetryCountMin {
		return false
	}
	if opts.TransactionGlobalTotalRetryCountMax > 0 && entry.TransactionGlobalTotalRetryCount > opts.TransactionGlobalTotalRetryCountMax {
		return false
	}
	if anyGlobalFamily := strings.TrimSpace(opts.TransactionGlobalAnyMatchFamily); anyGlobalFamily != "" && !containsFamily(entry.TransactionGlobalAnyMatchFamilies, anyGlobalFamily) {
		return false
	}
	if dominantGlobalFamily := strings.TrimSpace(opts.TransactionGlobalDominantFamily); dominantGlobalFamily != "" && entry.TransactionGlobalDominantFamily != dominantGlobalFamily {
		return false
	}
	if dominantFamily := strings.TrimSpace(opts.DominantFamily); dominantFamily != "" && entry.DominantFamily != dominantFamily {
		return false
	}
	if anyMatchFamily := strings.TrimSpace(opts.AnyMatchFamily); anyMatchFamily != "" && !containsFamily(entry.AnyMatchFamilies, anyMatchFamily) {
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

func sortDeadLetterEntries(entries []DeadLetterBacklogEntry, sortBy string) {
	sort.Slice(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]

		switch strings.TrimSpace(sortBy) {
		case "latest_dead_lettered_at":
			leftTime, leftOK := parseRFC3339(left.LatestDeadLetteredAt)
			rightTime, rightOK := parseRFC3339(right.LatestDeadLetteredAt)
			if leftOK && rightOK && !leftTime.Equal(rightTime) {
				return leftTime.After(rightTime)
			}
			if leftOK != rightOK {
				return leftOK
			}
		case "latest_manual_replay_at":
			leftTime, leftOK := parseRFC3339(left.LatestManualReplayAt)
			rightTime, rightOK := parseRFC3339(right.LatestManualReplayAt)
			if leftOK && rightOK && !leftTime.Equal(rightTime) {
				return leftTime.After(rightTime)
			}
			if leftOK != rightOK {
				return leftOK
			}
		default:
			if left.LatestRetryAttempt != right.LatestRetryAttempt {
				return left.LatestRetryAttempt > right.LatestRetryAttempt
			}
		}

		if left.LatestRetryAttempt != right.LatestRetryAttempt {
			return left.LatestRetryAttempt > right.LatestRetryAttempt
		}
		return left.TransactionReceiptID < right.TransactionReceiptID
	})
}

func subtypeFamily(subtype string) string {
	switch strings.TrimSpace(subtype) {
	case "retry-scheduled":
		return "retry"
	case "manual-retry-requested":
		return "manual-retry"
	case "dead-lettered":
		return "dead-letter"
	default:
		return ""
	}
}

func anyMatchFamilies(families map[string]struct{}) []string {
	if len(families) == 0 {
		return nil
	}

	values := make([]string, 0, len(families))
	for family := range families {
		values = append(values, family)
	}
	sort.Strings(values)
	return values
}

func dominantFamily(familyCounts map[string]int, latestRelevantFamily string) string {
	if len(familyCounts) == 0 {
		return ""
	}

	maxCount := 0
	candidates := make([]string, 0, len(familyCounts))
	for family, count := range familyCounts {
		switch {
		case count > maxCount:
			maxCount = count
			candidates = []string{family}
		case count == maxCount:
			candidates = append(candidates, family)
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}
	for _, family := range candidates {
		if family == latestRelevantFamily {
			return family
		}
	}

	sort.Strings(candidates)
	return candidates[0]
}

func containsFamily(families []string, want string) bool {
	for _, family := range families {
		if strings.EqualFold(family, want) {
			return true
		}
	}
	return false
}
