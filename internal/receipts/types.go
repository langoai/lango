package receipts

import "errors"

var (
	ErrSubmissionReceiptNotFound    = errors.New("submission receipt not found")
	ErrTransactionReceiptNotFound   = errors.New("transaction receipt not found")
	ErrInvalidSubmissionInput       = errors.New("invalid submission input")
	ErrInvalidReceiptEventType      = errors.New("invalid receipt event type")
	ErrInvalidPaymentApprovalStatus = errors.New("invalid payment approval status")
)

type ApprovalStatus string

const (
	ApprovalPending           ApprovalStatus = "pending"
	ApprovalApproved          ApprovalStatus = "approved"
	ApprovalRejected          ApprovalStatus = "rejected"
	ApprovalRevisionRequested ApprovalStatus = "revision-requested"
	ApprovalEscalated         ApprovalStatus = "escalated"
)

type SettlementStatus string

const (
	SettlementPending          SettlementStatus = "pending"
	SettlementPartiallySettled SettlementStatus = "partially-settled"
	SettlementSettled          SettlementStatus = "settled"
	SettlementDisputed         SettlementStatus = "disputed"
)

type PaymentApprovalStatus string

const (
	PaymentApprovalPending   PaymentApprovalStatus = "pending"
	PaymentApprovalApproved  PaymentApprovalStatus = "approved"
	PaymentApprovalRejected  PaymentApprovalStatus = "rejected"
	PaymentApprovalEscalated PaymentApprovalStatus = "escalated"
)

type EventType string

const (
	EventDraftExportability EventType = "draft_exportability"
	EventFinalExportability EventType = "final_exportability"
	EventApprovalRequested  EventType = "approval_requested"
	EventApprovalResolved   EventType = "approval_resolved"
	EventPaymentApproval    EventType = "payment_approval"
	EventSettlementUpdated  EventType = "settlement_updated"
	EventEscalated          EventType = "escalated"
	EventDisputed           EventType = "disputed"
)

type ProvenanceSummary struct {
	ReferenceID        string `json:"reference_id"`
	ConfigFingerprint  string `json:"config_fingerprint,omitempty"`
	SignerSummary      string `json:"signer_summary,omitempty"`
	AttributionSummary string `json:"attribution_summary,omitempty"`
}

type SubmissionReceipt struct {
	SubmissionReceiptID     string            `json:"submission_receipt_id"`
	TransactionReceiptID    string            `json:"transaction_receipt_id"`
	ArtifactLabel           string            `json:"artifact_label"`
	PayloadHash             string            `json:"payload_hash"`
	SourceLineageDigest     string            `json:"source_lineage_digest"`
	CanonicalApprovalStatus ApprovalStatus    `json:"canonical_approval_status"`
	CanonicalSettlementHint string            `json:"canonical_settlement_hint,omitempty"`
	ProvenanceSummary       ProvenanceSummary `json:"provenance_summary"`
}

type TransactionReceipt struct {
	TransactionReceiptID             string                `json:"transaction_receipt_id"`
	TransactionID                    string                `json:"transaction_id"`
	CurrentSubmissionReceiptID       string                `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus          ApprovalStatus        `json:"canonical_approval_status"`
	CanonicalSettlementStatus        SettlementStatus      `json:"canonical_settlement_status"`
	CurrentPaymentApprovalStatus     PaymentApprovalStatus `json:"current_payment_approval_status"`
	CanonicalPaymentApprovalDecision string                `json:"canonical_payment_approval_decision,omitempty"`
	CanonicalPaymentSettlementHint   string                `json:"canonical_payment_settlement_hint,omitempty"`
}
