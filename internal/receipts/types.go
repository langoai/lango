package receipts

import "errors"

var (
	ErrSubmissionReceiptNotFound            = errors.New("submission receipt not found")
	ErrTransactionReceiptNotFound           = errors.New("transaction receipt not found")
	ErrInvalidSubmissionInput               = errors.New("invalid submission input")
	ErrInvalidReceiptEventType              = errors.New("invalid receipt event type")
	ErrInvalidPaymentApprovalStatus         = errors.New("invalid payment approval status")
	ErrInvalidEscrowExecutionStatus         = errors.New("invalid escrow execution status")
	ErrInvalidEscrowExecutionState          = errors.New("invalid escrow execution state")
	ErrInvalidKnowledgeExchangeRuntimeState = errors.New("invalid knowledge exchange runtime state")
	ErrInvalidSettlementProgressionState    = errors.New("invalid settlement progression state")
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

type EscrowExecutionStatus string

const (
	EscrowExecutionStatusPending EscrowExecutionStatus = "pending"
	EscrowExecutionStatusCreated EscrowExecutionStatus = "created"
	EscrowExecutionStatusFunded  EscrowExecutionStatus = "funded"
	EscrowExecutionStatusFailed  EscrowExecutionStatus = "failed"
)

type PaymentApprovalStatus string

const (
	PaymentApprovalPending   PaymentApprovalStatus = "pending"
	PaymentApprovalApproved  PaymentApprovalStatus = "approved"
	PaymentApprovalRejected  PaymentApprovalStatus = "rejected"
	PaymentApprovalEscalated PaymentApprovalStatus = "escalated"
)

type KnowledgeExchangeRuntimeStatus string

const (
	RuntimeStatusOpened                KnowledgeExchangeRuntimeStatus = "opened"
	RuntimeStatusExportabilityAdvisory KnowledgeExchangeRuntimeStatus = "exportability-advisory"
	RuntimeStatusPaymentApproved       KnowledgeExchangeRuntimeStatus = "payment-approved"
	RuntimeStatusPaymentAuthorized     KnowledgeExchangeRuntimeStatus = "payment-authorized"
	RuntimeStatusEscrowFunded          KnowledgeExchangeRuntimeStatus = "escrow-funded"
	RuntimeStatusWorkStarted           KnowledgeExchangeRuntimeStatus = "work-started"
	RuntimeStatusSubmissionReceived    KnowledgeExchangeRuntimeStatus = "submission-received"
	RuntimeStatusReleaseApproved       KnowledgeExchangeRuntimeStatus = "release-approved"
	RuntimeStatusRevisionRequested     KnowledgeExchangeRuntimeStatus = "revision-requested"
	RuntimeStatusEscalated             KnowledgeExchangeRuntimeStatus = "escalated"
	RuntimeStatusDisputeReady          KnowledgeExchangeRuntimeStatus = "dispute-ready"
)

type SettlementProgressionStatus string

const (
	SettlementProgressionPending               SettlementProgressionStatus = "pending"
	SettlementProgressionInProgress            SettlementProgressionStatus = "in-progress"
	SettlementProgressionReviewNeeded          SettlementProgressionStatus = "review-needed"
	SettlementProgressionApprovedForSettlement SettlementProgressionStatus = "approved-for-settlement"
	SettlementProgressionPartiallySettled      SettlementProgressionStatus = "partially-settled"
	SettlementProgressionSettled               SettlementProgressionStatus = "settled"
	SettlementProgressionDisputeReady          SettlementProgressionStatus = "dispute-ready"
)

type SettlementProgressionReasonCode string

const (
	SettlementProgressionReasonCodeApprove         SettlementProgressionReasonCode = "approve"
	SettlementProgressionReasonCodeReject          SettlementProgressionReasonCode = "reject"
	SettlementProgressionReasonCodeRequestRevision SettlementProgressionReasonCode = "request-revision"
	SettlementProgressionReasonCodeEscalate        SettlementProgressionReasonCode = "escalate"
)

type EventType string

const (
	EventDraftExportability         EventType = "draft_exportability"
	EventFinalExportability         EventType = "final_exportability"
	EventApprovalRequested          EventType = "approval_requested"
	EventApprovalResolved           EventType = "approval_resolved"
	EventPaymentApproval            EventType = "payment_approval"
	EventPaymentExecutionAuthorized EventType = "payment_execution_authorized"
	EventPaymentExecutionDenied     EventType = "payment_execution_denied"
	EventEscrowExecutionStarted     EventType = "escrow_execution_started"
	EventEscrowExecutionCreated     EventType = "escrow_execution_created"
	EventEscrowExecutionFunded      EventType = "escrow_execution_funded"
	EventEscrowExecutionFailed      EventType = "escrow_execution_failed"
	EventSettlementUpdated          EventType = "settlement_updated"
	EventSettlementExecutionFailed  EventType = "settlement_execution_failed"
	EventEscalated                  EventType = "escalated"
	EventDisputed                   EventType = "disputed"
)

type EscrowMilestoneInput struct {
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

type EscrowExecutionInput struct {
	BuyerDID   string                 `json:"buyer_did"`
	SellerDID  string                 `json:"seller_did"`
	Amount     string                 `json:"amount"`
	Reason     string                 `json:"reason"`
	TaskID     string                 `json:"task_id,omitempty"`
	Milestones []EscrowMilestoneInput `json:"milestones"`
}

type ProvenanceSummary struct {
	ReferenceID        string `json:"reference_id"`
	ConfigFingerprint  string `json:"config_fingerprint,omitempty"`
	SignerSummary      string `json:"signer_summary,omitempty"`
	AttributionSummary string `json:"attribution_summary,omitempty"`
}

type OpenTransactionInput struct {
	TransactionID  string `json:"transaction_id"`
	Counterparty   string `json:"counterparty"`
	RequestedScope string `json:"requested_scope"`
	PriceContext   string `json:"price_context"`
	TrustContext   string `json:"trust_context"`
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
	TransactionReceiptID            string                          `json:"transaction_receipt_id"`
	TransactionID                   string                          `json:"transaction_id"`
	Counterparty                    string                          `json:"counterparty,omitempty"`
	RequestedScope                  string                          `json:"requested_scope,omitempty"`
	PriceContext                    string                          `json:"price_context,omitempty"`
	TrustContext                    string                          `json:"trust_context,omitempty"`
	KnowledgeExchangeRuntimeStatus  KnowledgeExchangeRuntimeStatus  `json:"knowledge_exchange_runtime_status,omitempty"`
	SettlementProgressionStatus     SettlementProgressionStatus     `json:"settlement_progression_status,omitempty"`
	SettlementProgressionReasonCode SettlementProgressionReasonCode `json:"settlement_progression_reason_code,omitempty"`
	SettlementProgressionReason     string                          `json:"settlement_progression_reason,omitempty"`
	PartialSettlementHint           string                          `json:"partial_settlement_hint,omitempty"`
	DisputeReady                    bool                            `json:"dispute_ready,omitempty"`
	CurrentSubmissionReceiptID      string                          `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus         ApprovalStatus                  `json:"canonical_approval_status"`
	CanonicalSettlementStatus       SettlementStatus                `json:"canonical_settlement_status"`
	CurrentPaymentApprovalStatus    PaymentApprovalStatus           `json:"current_payment_approval_status"`
	CanonicalDecision               string                          `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint         string                          `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus           EscrowExecutionStatus           `json:"escrow_execution_status,omitempty"`
	EscrowReference                 string                          `json:"escrow_reference,omitempty"`
	EscrowExecutionInput            *EscrowExecutionInput           `json:"escrow_execution_input,omitempty"`
}

type SettlementCloseoutRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	ResolvedAmount       string `json:"resolved_amount"`
	RuntimeReference     string `json:"runtime_reference,omitempty"`
}

type SettlementFailureRequest struct {
	TransactionReceiptID string `json:"transaction_receipt_id"`
	SubmissionReceiptID  string `json:"submission_receipt_id"`
	ResolvedAmount       string `json:"resolved_amount,omitempty"`
	Reason               string `json:"reason"`
}
