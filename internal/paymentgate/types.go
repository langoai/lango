package paymentgate

type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

type DenyReason string

const (
	DenyReasonMissingReceipt        DenyReason = "missing_receipt"
	DenyReasonApprovalNotApproved   DenyReason = "approval_not_approved"
	DenyReasonExecutionModeMismatch DenyReason = "execution_mode_mismatch"
)

type Request struct {
	TransactionReceiptID string
}

type Result struct {
	Decision   Decision   `json:"decision"`
	DenyReason DenyReason `json:"deny_reason,omitempty"`
}
