package partialsettlementexecution

import "github.com/langoai/lango/internal/receipts"

type DenyReason string

const (
	DenyReasonMissingReceipt           DenyReason = "missing_receipt"
	DenyReasonNoCurrentSubmission      DenyReason = "no_current_submission"
	DenyReasonNotApprovedForSettlement DenyReason = "not_approved_for_settlement"
	DenyReasonPartialHintMissing       DenyReason = "partial_hint_missing"
	DenyReasonPartialHintInvalid       DenyReason = "partial_hint_invalid"
	DenyReasonAlreadyPartiallySettled  DenyReason = "already_partially_settled"
)

type FailureKind string

const (
	FailureKindDenied          FailureKind = "denied"
	FailureKindExecutionFailed FailureKind = "execution-failed"
)

type ResultStatus string

const (
	ResultStatusDenied                 ResultStatus = "denied"
	ResultStatusFailed                 ResultStatus = "execution-failure"
	ResultStatusPartiallySettledTarget ResultStatus = "partially-settled-target"
)

const (
	StatusDenied                 = ResultStatusDenied
	StatusFailed                 = ResultStatusFailed
	StatusPartiallySettledTarget = ResultStatusPartiallySettledTarget
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
	ExecutedAmount              string                               `json:"executed_amount,omitempty"`
	RemainingAmount             string                               `json:"remaining_amount,omitempty"`
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

type DirectPaymentRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	Counterparty         string `json:"counterparty,omitempty"`
	Amount               string `json:"amount"`
}

type DirectPaymentResult struct {
	Reference string `json:"reference,omitempty"`
}
