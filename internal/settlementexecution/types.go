package settlementexecution

import "github.com/langoai/lango/internal/receipts"

type DenyReason string

const (
	DenyReasonMissingReceipt           DenyReason = "missing receipt"
	DenyReasonNoCurrentSubmission      DenyReason = "no current submission"
	DenyReasonNotApprovedForSettlement DenyReason = "not approved-for-settlement"
	DenyReasonAmountUnresolved         DenyReason = "amount unresolved"
)

type FailureKind string

const (
	FailureKindDenied          FailureKind = "denied"
	FailureKindExecutionFailed FailureKind = "execution-failed"
)

type ResultStatus string

const (
	ResultStatusDenied        ResultStatus = "denied"
	ResultStatusFailed        ResultStatus = "failed"
	ResultStatusSettledTarget ResultStatus = "settled-target"
)

const (
	StatusDenied        = ResultStatusDenied
	StatusFailed        = ResultStatusFailed
	StatusSettledTarget = ResultStatusSettledTarget
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

type DirectPaymentRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	Counterparty         string `json:"counterparty,omitempty"`
	Amount               string `json:"amount"`
}

type DirectPaymentResult struct {
	Reference string `json:"reference,omitempty"`
}

type closeoutRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	ResolvedAmount       string `json:"resolved_amount"`
	RuntimeReference     string `json:"runtime_reference,omitempty"`
}

type CloseoutRequest = closeoutRequest

type failureRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	ResolvedAmount       string `json:"resolved_amount,omitempty"`
	Reason               string `json:"reason"`
}

type FailureRecordRequest = failureRequest
