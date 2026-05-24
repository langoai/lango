package escrowrefund

import "github.com/langoai/lango/internal/receipts"

type DenyReason string

const (
	DenyReasonMissingReceipt       DenyReason = "missing_receipt"
	DenyReasonNoCurrentSubmission  DenyReason = "no_current_submission"
	DenyReasonEscrowNotFunded      DenyReason = "escrow_not_funded"
	DenyReasonNotReviewNeeded      DenyReason = "not_review_needed"
	DenyReasonAdjudicationMissing  DenyReason = "adjudication_missing"
	DenyReasonAdjudicationMismatch DenyReason = "adjudication_mismatch"
	DenyReasonAmountUnresolved     DenyReason = "amount_unresolved"
)

type FailureKind string

const (
	FailureKindDenied          FailureKind = "denied"
	FailureKindExecutionFailed FailureKind = "execution-failed"
)

type ResultStatus string

const (
	ResultStatusDenied         ResultStatus = "denied"
	ResultStatusFailed         ResultStatus = "execution-failure"
	ResultStatusRefundExecuted ResultStatus = "refund-executed"
)

const (
	StatusDenied         = ResultStatusDenied
	StatusFailed         = ResultStatusFailed
	StatusRefundExecuted = ResultStatusRefundExecuted
)

type ExecuteRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
}

type Request = ExecuteRequest

type Result struct {
	Status                      ResultStatus                         `json:"status"`
	TransactionReceiptID        string                               `json:"transaction_receipt_id,omitempty"`
	SubmissionReceiptID         string                               `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus receipts.SettlementProgressionStatus `json:"settlement_progression_status,omitempty"`
	ResolvedAmount              string                               `json:"resolved_amount,omitempty"`
	RuntimeReference            string                               `json:"runtime_reference,omitempty"`
	Failure                     *Failure                             `json:"failure,omitempty"`
}

type Failure struct {
	Kind       FailureKind `json:"kind"`
	DenyReason DenyReason  `json:"deny_reason,omitempty"`
	Message    string      `json:"message"`
}

type ExecutionError struct {
	Kind       FailureKind
	DenyReason DenyReason
	Message    string
	Err        error
}

func (e *ExecutionError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *ExecutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type RefundRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	EscrowReference      string `json:"escrow_reference,omitempty"`
	Amount               string `json:"amount"`
}

type RefundResult struct {
	Reference string `json:"reference,omitempty"`
}
