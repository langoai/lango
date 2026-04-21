package escrowexecution

import "github.com/langoai/lango/internal/receipts"

type Request struct {
	TransactionReceiptID string
	SubmissionReceiptID  string
}

type Result struct {
	TransactionReceiptID string
	SubmissionReceiptID  string
	EscrowID             string
	Status               receipts.EscrowExecutionStatus
}
