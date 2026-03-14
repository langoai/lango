// Package bundler provides an ERC-4337 bundler JSON-RPC client.
package bundler

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Sentinel errors for bundler operations.
var (
	ErrInvalidUserOp = errors.New("invalid user operation")
	ErrBundlerError  = errors.New("bundler RPC error")
)

// UserOperation represents an ERC-4337 UserOperation for the bundler client.
// This is a bundler-local mirror of the parent smartaccount.UserOperation
// to avoid import cycles.
type UserOperation struct {
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

// UserOpResult contains the result of submitting a UserOp.
type UserOpResult struct {
	UserOpHash common.Hash `json:"userOpHash"`
	TxHash     common.Hash `json:"txHash,omitempty"`
	Success    bool        `json:"success"`
	GasUsed    uint64      `json:"gasUsed,omitempty"`
}

// GasEstimate contains gas estimation for a UserOp.
type GasEstimate struct {
	CallGasLimit         *big.Int `json:"callGasLimit"`
	VerificationGasLimit *big.Int `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int `json:"preVerificationGas"`
}

// GasFees contains EIP-1559 gas fee parameters.
type GasFees struct {
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
}

// jsonrpcRequest is a JSON-RPC 2.0 request.
type jsonrpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// jsonrpcResponse is a JSON-RPC 2.0 response.
type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
