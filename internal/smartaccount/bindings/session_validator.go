package bindings

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// SessionValidatorABI is the ABI for LangoSessionValidator.
const SessionValidatorABI = `[
	{
		"inputs": [
			{"name": "sessionKey", "type": "address"},
			{
				"components": [
					{"name": "allowedTargets", "type": "address[]"},
					{"name": "allowedFunctions", "type": "bytes4[]"},
					{"name": "spendLimit", "type": "uint256"},
					{"name": "spentAmount", "type": "uint256"},
					{"name": "validAfter", "type": "uint48"},
					{"name": "validUntil", "type": "uint48"},
					{"name": "active", "type": "bool"}
				],
				"name": "policy",
				"type": "tuple"
			}
		],
		"name": "registerSessionKey",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "sessionKey", "type": "address"}
		],
		"name": "revokeSessionKey",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "sessionKey", "type": "address"}
		],
		"name": "getSessionKeyPolicy",
		"outputs": [
			{
				"components": [
					{"name": "allowedTargets", "type": "address[]"},
					{"name": "allowedFunctions", "type": "bytes4[]"},
					{"name": "spendLimit", "type": "uint256"},
					{"name": "spentAmount", "type": "uint256"},
					{"name": "validAfter", "type": "uint48"},
					{"name": "validUntil", "type": "uint48"},
					{"name": "active", "type": "bool"}
				],
				"name": "",
				"type": "tuple"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "sessionKey", "type": "address"}
		],
		"name": "isSessionKeyActive",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// SessionValidatorClient provides typed access to
// the LangoSessionValidator contract.
type SessionValidatorClient struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewSessionValidatorClient creates a new session validator client.
func NewSessionValidatorClient(
	caller contract.ContractCaller,
	address common.Address,
	chainID int64,
) *SessionValidatorClient {
	return &SessionValidatorClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: SessionValidatorABI,
	}
}

// RegisterSessionKey registers a new session key with its policy.
func (c *SessionValidatorClient) RegisterSessionKey(
	ctx context.Context,
	sessionKey common.Address,
	policy interface{},
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "registerSessionKey",
			Args:    []interface{}{sessionKey, policy},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"register session key: %w", err,
		)
	}
	return result.TxHash, nil
}

// RevokeSessionKey revokes an existing session key.
func (c *SessionValidatorClient) RevokeSessionKey(
	ctx context.Context,
	sessionKey common.Address,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "revokeSessionKey",
			Args:    []interface{}{sessionKey},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"revoke session key: %w", err,
		)
	}
	return result.TxHash, nil
}

// GetSessionKeyPolicy retrieves the policy for a session key.
func (c *SessionValidatorClient) GetSessionKeyPolicy(
	ctx context.Context,
	sessionKey common.Address,
) (interface{}, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getSessionKeyPolicy",
			Args:    []interface{}{sessionKey},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get session key policy: %w", err,
		)
	}
	if len(result.Data) > 0 {
		return result.Data[0], nil
	}
	return nil, nil
}

// IsSessionKeyActive checks whether a session key is active.
func (c *SessionValidatorClient) IsSessionKeyActive(
	ctx context.Context,
	sessionKey common.Address,
) (bool, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "isSessionKeyActive",
			Args:    []interface{}{sessionKey},
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"check session key: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(bool); ok {
			return v, nil
		}
	}
	return false, nil
}
