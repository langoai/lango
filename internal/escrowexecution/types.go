package escrowexecution

import "github.com/langoai/lango/internal/receipts"

type Request struct {
	TransactionReceiptID string
}

type Result struct {
	TransactionReceiptID  string
	SubmissionReceiptID   string
	EscrowReference       string
	EscrowExecutionStatus receipts.EscrowExecutionStatus
}
