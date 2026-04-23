package escrowadjudication

import "github.com/langoai/lango/internal/receipts"

type Outcome string

const (
	OutcomeRelease Outcome = "release"
	OutcomeRefund  Outcome = "refund"
)

type DenyReason string

const (
	DenyReasonMissingReceipt      DenyReason = "missing_receipt"
	DenyReasonNoCurrentSubmission DenyReason = "no_current_submission"
	DenyReasonEscrowNotFunded     DenyReason = "escrow_not_funded"
	DenyReasonNotDisputeReady     DenyReason = "not_dispute_ready"
	DenyReasonHoldEvidenceMissing DenyReason = "hold_evidence_missing"
	DenyReasonInvalidOutcome      DenyReason = "invalid_outcome"
)

type FailureKind string

const (
	FailureKindDenied      FailureKind = "denied"
	FailureKindApplyFailed FailureKind = "apply-failed"
)

type ResultStatus string

const (
	ResultStatusDenied      ResultStatus = "denied"
	ResultStatusFailed      ResultStatus = "failed"
	ResultStatusAdjudicated ResultStatus = "adjudicated"
)

const (
	StatusDenied      = ResultStatusDenied
	StatusFailed      = ResultStatusFailed
	StatusAdjudicated = ResultStatusAdjudicated
)

type AdjudicateRequest struct {
	TransactionReceiptID string  `json:"transaction_receipt_id"`
	Outcome              Outcome `json:"outcome"`
	Reason               string  `json:"reason,omitempty"`
}

type Request = AdjudicateRequest

type Result struct {
	Status                      ResultStatus                         `json:"status"`
	TransactionReceiptID        string                               `json:"transaction_receipt_id,omitempty"`
	SubmissionReceiptID         string                               `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus receipts.SettlementProgressionStatus `json:"settlement_progression_status,omitempty"`
	EscrowReference             string                               `json:"escrow_reference,omitempty"`
	Outcome                     Outcome                              `json:"outcome,omitempty"`
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
