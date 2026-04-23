package postadjudicationstatus

import (
	"context"
	"errors"

	"github.com/langoai/lango/internal/receipts"
)

var (
	ErrTransactionReceiptNotFound = errors.New("transaction receipt not found")
	ErrCurrentSubmissionMissing   = errors.New("current submission missing")
)

type DeadLetterBacklogEntry struct {
	TransactionReceiptID    string `json:"transaction_receipt_id"`
	SubmissionReceiptID     string `json:"submission_receipt_id"`
	Adjudication            string `json:"adjudication"`
	LatestDeadLetterReason  string `json:"latest_dead_letter_reason,omitempty"`
	LatestRetryAttempt      int    `json:"latest_retry_attempt,omitempty"`
	LatestDispatchReference string `json:"latest_dispatch_reference,omitempty"`
}

type CanonicalSnapshot struct {
	TransactionReceipt receipts.TransactionReceipt `json:"transaction_receipt"`
	SubmissionReceipt  receipts.SubmissionReceipt  `json:"submission_receipt"`
	SubmissionEvents   []receipts.ReceiptEvent     `json:"submission_events,omitempty"`
}

type RetryDeadLetterSummary struct {
	HasDeadLetter           bool   `json:"has_dead_letter"`
	LatestDeadLetterReason  string `json:"latest_dead_letter_reason,omitempty"`
	LatestRetryAttempt      int    `json:"latest_retry_attempt,omitempty"`
	LatestDispatchReference string `json:"latest_dispatch_reference,omitempty"`
	LatestStatusSubtype     string `json:"latest_status_subtype,omitempty"`
}

type TransactionStatus struct {
	CanonicalSnapshot      CanonicalSnapshot      `json:"canonical_snapshot"`
	RetryDeadLetterSummary RetryDeadLetterSummary `json:"retry_dead_letter_summary"`
}

type receiptStore interface {
	ListTransactionReceipts(context.Context) ([]receipts.TransactionReceipt, error)
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
}
