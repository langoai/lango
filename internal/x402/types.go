// Package x402 implements the X402 HTTP payment protocol for blockchain micropayments.
// When a server responds with HTTP 402 Payment Required, this package parses
// the payment challenge and coordinates with the payment service to complete it.
package x402

// Challenge represents a parsed HTTP 402 payment challenge from response headers.
type Challenge struct {
	// PaymentURL is the URL that returned the 402 status.
	PaymentURL string `json:"paymentUrl"`

	// Amount is the requested payment amount (e.g. "0.01").
	Amount string `json:"amount"`

	// TokenAddress is the ERC-20 token contract address.
	TokenAddress string `json:"tokenAddress"`

	// RecipientAddress is the payment recipient's wallet address.
	RecipientAddress string `json:"recipientAddress"`

	// Network is the blockchain network (e.g. "base-sepolia").
	Network string `json:"network"`

	// ChainID is the numeric chain ID (derived from Network or header).
	ChainID int64 `json:"chainId,omitempty"`
}

// PaymentPayload is the signed payment proof sent back to the server
// in the X-PAYMENT header to retry the request.
type PaymentPayload struct {
	// TxHash is the on-chain transaction hash proving payment.
	TxHash string `json:"txHash"`

	// From is the sender wallet address.
	From string `json:"from"`

	// ChainID is the chain the payment was made on.
	ChainID int64 `json:"chainId"`
}

// Header constants for X402 protocol.
const (
	// HeaderPaymentAmount is the header specifying the required payment amount.
	HeaderPaymentAmount = "X-Payment-Amount"

	// HeaderPaymentToken is the header specifying the token contract address.
	HeaderPaymentToken = "X-Payment-Token"

	// HeaderPaymentRecipient is the header specifying the payment recipient.
	HeaderPaymentRecipient = "X-Payment-Recipient"

	// HeaderPaymentNetwork is the header specifying the blockchain network.
	HeaderPaymentNetwork = "X-Payment-Network"

	// HeaderPaymentChainID is the header specifying the chain ID.
	HeaderPaymentChainID = "X-Payment-ChainId"

	// HeaderPayment is the header used to send payment proof on retry.
	HeaderPayment = "X-PAYMENT"
)
