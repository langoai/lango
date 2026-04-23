package disputehold

import "github.com/langoai/lango/internal/receipts"

type DenyReason string

const (
	DenyReasonMissingReceipt         DenyReason = "missing_receipt"
	DenyReasonNoCurrentSubmission    DenyReason = "no_current_submission"
	DenyReasonEscrowNotFunded        DenyReason = "escrow_not_funded"
	DenyReasonNotDisputeReady        DenyReason = "not_dispute_ready"
	DenyReasonEscrowReferenceMissing DenyReason = "escrow_reference_missing"
)

type FailureKind string

const (
	FailureKindDenied          FailureKind = "denied"
	FailureKindExecutionFailed FailureKind = "execution-failed"
)

type ResultStatus string

const (
	ResultStatusDenied      ResultStatus = "denied"
	ResultStatusFailed      ResultStatus = "execution-failure"
	ResultStatusHoldApplied ResultStatus = "hold-applied"
)

const (
	StatusDenied      = ResultStatusDenied
	StatusFailed      = ResultStatusFailed
	StatusHoldApplied = ResultStatusHoldApplied
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
	EscrowReference             string                               `json:"escrow_reference,omitempty"`
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

type EscrowHoldRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	EscrowReference      string `json:"escrow_reference"`
}

type HoldRequest = EscrowHoldRequest

type HoldResult struct {
	Reference string `json:"reference,omitempty"`
}
