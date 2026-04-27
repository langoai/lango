package knowledgeruntime

import "github.com/langoai/lango/internal/receipts"

type OpenTransactionRequest struct {
	TransactionID  string
	Counterparty   string
	RequestedScope string
	PriceContext   string
	TrustContext   string
}

type OpenTransactionResult struct {
	TransactionReceiptID string
	RuntimeStatus        receipts.KnowledgeExchangeRuntimeStatus
}

type Branch string

const (
	BranchPrepay Branch = "prepay"
	BranchEscrow Branch = "escrow"
)

type BranchSelection struct {
	TransactionReceiptID string
	Branch               Branch
}
