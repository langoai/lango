package settlementprogression

import (
	"errors"

	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/receipts"
)

var (
	ErrInvalidApplyReleaseOutcomeRequest = errors.New("invalid apply release outcome request")
	ErrUnsupportedReleaseDecision        = errors.New("unsupported release decision")
)

type ReleaseOutcome struct {
	Decision approvalflow.Decision `json:"decision"`
	Reason   string                `json:"reason,omitempty"`
}

type ApplyReleaseOutcomeRequest struct {
	TransactionReceiptID string         `json:"transaction_receipt_id"`
	Outcome              ReleaseOutcome `json:"outcome"`
}

type ApplyReleaseOutcomeResult struct {
	Transaction receipts.TransactionReceipt `json:"transaction"`
	Outcome     SettlementOutcome           `json:"outcome"`
}

type SettlementOutcome struct {
	ProgressionStatus     receipts.SettlementProgressionStatus     `json:"progression_status"`
	ProgressionReasonCode receipts.SettlementProgressionReasonCode `json:"progression_reason_code,omitempty"`
	ProgressionReason     string                                   `json:"progression_reason,omitempty"`
	PartialHint           string                                   `json:"partial_hint,omitempty"`
}
