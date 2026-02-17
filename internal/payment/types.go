// Package payment implements blockchain payment services for USDC on EVM chains.
package payment

import "time"

// PaymentRequest describes a payment to be sent.
type PaymentRequest struct {
	// To is the recipient wallet address.
	To string `json:"to"`

	// Amount is the USDC amount as a decimal string (e.g. "1.50").
	Amount string `json:"amount"`

	// Purpose is a human-readable description of the payment.
	Purpose string `json:"purpose,omitempty"`

	// SessionKey is the agent session that initiated the payment.
	SessionKey string `json:"sessionKey,omitempty"`

	// X402URL is the URL that triggered an X402 payment (if applicable).
	X402URL string `json:"x402Url,omitempty"`
}

// PaymentReceipt is returned after a payment is submitted.
type PaymentReceipt struct {
	TxHash    string    `json:"txHash"`
	Status    string    `json:"status"`
	Amount    string    `json:"amount"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	ChainID   int64     `json:"chainId"`
	Timestamp time.Time `json:"timestamp"`
}

// TransactionInfo combines a receipt with contextual information.
type TransactionInfo struct {
	TxHash       string    `json:"txHash,omitempty"`
	Status       string    `json:"status"`
	Amount       string    `json:"amount"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	ChainID      int64     `json:"chainId"`
	Purpose      string    `json:"purpose,omitempty"`
	X402URL      string    `json:"x402Url,omitempty"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// X402Challenge represents a parsed HTTP 402 payment challenge.
type X402Challenge struct {
	PaymentURL       string `json:"paymentUrl"`
	Amount           string `json:"amount"`
	TokenAddress     string `json:"tokenAddress"`
	RecipientAddress string `json:"recipientAddress"`
	Network          string `json:"network"`
}
