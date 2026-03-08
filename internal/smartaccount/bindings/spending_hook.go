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
			{"name": "perTxLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "cumulativeLimit", "type": "uint256"}
		],
		"name": "setLimits",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "account", "type": "address"}
		],
		"name": "getConfig",
		"outputs": [
			{"name": "perTxLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "cumulativeLimit", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "account", "type": "address"},
			{"name": "sessionKey", "type": "address"}
		],
		"name": "getSpendState",
		"outputs": [
			{"name": "dailySpent", "type": "uint256"},
			{"name": "cumulativeSpent", "type": "uint256"},
			{"name": "lastResetDay", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

// SpendingConfig represents the spending limits for an account.
type SpendingConfig struct {
	PerTxLimit      *big.Int
	DailyLimit      *big.Int
	CumulativeLimit *big.Int
}

// SpendState represents the current spending state for a session.
type SpendState struct {
	DailySpent      *big.Int
	CumulativeSpent *big.Int
	LastResetDay    *big.Int
}

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

// SetLimits configures the spending limits for the caller's account.
func (c *SpendingHookClient) SetLimits(
	ctx context.Context,
	perTxLimit, dailyLimit, cumulativeLimit *big.Int,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "setLimits",
			Args:    []interface{}{perTxLimit, dailyLimit, cumulativeLimit},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"set limits: %w", err,
		)
	}
	return result.TxHash, nil
}

// GetConfig retrieves the spending limits for an account.
func (c *SpendingHookClient) GetConfig(
	ctx context.Context,
	account common.Address,
) (*SpendingConfig, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getConfig",
			Args:    []interface{}{account},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get config: %w", err,
		)
	}
	if len(result.Data) >= 3 {
		config := &SpendingConfig{}
		if v, ok := result.Data[0].(*big.Int); ok {
			config.PerTxLimit = v
		}
		if v, ok := result.Data[1].(*big.Int); ok {
			config.DailyLimit = v
		}
		if v, ok := result.Data[2].(*big.Int); ok {
			config.CumulativeLimit = v
		}
		return config, nil
	}
	return &SpendingConfig{
		PerTxLimit:      big.NewInt(0),
		DailyLimit:      big.NewInt(0),
		CumulativeLimit: big.NewInt(0),
	}, nil
}

// GetSpendState retrieves the spending state for an account's session key.
func (c *SpendingHookClient) GetSpendState(
	ctx context.Context,
	account, sessionKey common.Address,
) (*SpendState, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getSpendState",
			Args:    []interface{}{account, sessionKey},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get spend state: %w", err,
		)
	}
	if len(result.Data) >= 3 {
		state := &SpendState{}
		if v, ok := result.Data[0].(*big.Int); ok {
			state.DailySpent = v
		}
		if v, ok := result.Data[1].(*big.Int); ok {
			state.CumulativeSpent = v
		}
		if v, ok := result.Data[2].(*big.Int); ok {
			state.LastResetDay = v
		}
		return state, nil
	}
	return &SpendState{
		DailySpent:      big.NewInt(0),
		CumulativeSpent: big.NewInt(0),
		LastResetDay:    big.NewInt(0),
	}, nil
}
