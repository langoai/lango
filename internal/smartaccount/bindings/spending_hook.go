package bindings

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// SpendingHookABI is the ABI for LangoSpendingHook.
const SpendingHookABI = `[
	{
		"inputs": [
			{"name": "account", "type": "address"}
		],
		"name": "getSpentAmount",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "account", "type": "address"}
		],
		"name": "getLimit",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "account", "type": "address"}
		],
		"name": "resetSpentAmount",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "account", "type": "address"},
			{"name": "limit", "type": "uint256"}
		],
		"name": "setLimit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// SpendingHookClient provides typed access to the
// LangoSpendingHook contract.
type SpendingHookClient struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewSpendingHookClient creates a new spending hook client.
func NewSpendingHookClient(
	caller contract.ContractCaller,
	address common.Address,
	chainID int64,
) *SpendingHookClient {
	return &SpendingHookClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: SpendingHookABI,
	}
}

// GetSpentAmount returns the amount spent by an account.
func (c *SpendingHookClient) GetSpentAmount(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getSpentAmount",
			Args:    []interface{}{account},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get spent amount: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(*big.Int); ok {
			return v, nil
		}
	}
	return big.NewInt(0), nil
}

// GetLimit returns the spending limit for an account.
func (c *SpendingHookClient) GetLimit(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getLimit",
			Args:    []interface{}{account},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get limit: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(*big.Int); ok {
			return v, nil
		}
	}
	return big.NewInt(0), nil
}

// ResetSpentAmount resets the spent amount for an account.
func (c *SpendingHookClient) ResetSpentAmount(
	ctx context.Context,
	account common.Address,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "resetSpentAmount",
			Args:    []interface{}{account},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"reset spent amount: %w", err,
		)
	}
	return result.TxHash, nil
}

// SetLimit sets the spending limit for an account.
func (c *SpendingHookClient) SetLimit(
	ctx context.Context,
	account common.Address,
	limit *big.Int,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "setLimit",
			Args:    []interface{}{account, limit},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"set limit: %w", err,
		)
	}
	return result.TxHash, nil
}
