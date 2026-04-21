package paymentgate

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
)

type DenyReason string

const (
	ReasonMissingReceipt        DenyReason = "missing_receipt"
	ReasonApprovalNotApproved   DenyReason = "approval_not_approved"
	ReasonStaleState            DenyReason = "stale_state"
	ReasonExecutionModeMismatch DenyReason = "execution_mode_mismatch"
)

type Request struct {
	TransactionReceiptID string
	SubmissionReceiptID  string
	ToolName             string
	Context              map[string]interface{}
}

type Result struct {
	Decision Decision   `json:"decision"`
	Reason   DenyReason `json:"reason,omitempty"`
}
