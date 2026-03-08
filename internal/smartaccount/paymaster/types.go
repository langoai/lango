// Package paymaster provides ERC-4337 paymaster integration for gasless transactions.
package paymaster

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PaymasterProvider sponsors UserOperations via a paymaster service.
type PaymasterProvider interface {
	SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error)
	Type() string
}

// SponsorRequest contains UserOp data for paymaster sponsorship.
type SponsorRequest struct {
	UserOp     *UserOpData        `json:"userOp"`
	EntryPoint common.Address     `json:"entryPoint"`
	ChainID    int64              `json:"chainId"`
	Stub       bool               `json:"stub"`
	Context    map[string]string  `json:"context,omitempty"`
}

// SponsorResult contains paymaster response data.
type SponsorResult struct {
	PaymasterAndData []byte       `json:"paymasterAndData"`
	GasOverrides     *GasOverrides `json:"gasOverrides,omitempty"`
}

// GasOverrides allows the paymaster to override gas estimates.
type GasOverrides struct {
	CallGasLimit         *big.Int `json:"callGasLimit,omitempty"`
	VerificationGasLimit *big.Int `json:"verificationGasLimit,omitempty"`
	PreVerificationGas   *big.Int `json:"preVerificationGas,omitempty"`
}

// UserOpData is a paymaster-local UserOp mirror to avoid import cycles.
type UserOpData struct {
	Sender               common.Address `json:"sender"`
	Nonce                *big.Int       `json:"nonce"`
	InitCode             []byte         `json:"initCode"`
	CallData             []byte         `json:"callData"`
	CallGasLimit         *big.Int       `json:"callGasLimit"`
	VerificationGasLimit *big.Int       `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int       `json:"preVerificationGas"`
	MaxFeePerGas         *big.Int       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int       `json:"maxPriorityFeePerGas"`
	PaymasterAndData     []byte         `json:"paymasterAndData"`
	Signature            []byte         `json:"signature"`
}
