package bindings

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// Safe7579ABI is the ABI for the Safe7579 adapter contract.
const Safe7579ABI = `[
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "initData", "type": "bytes"}
		],
		"name": "installModule",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "deInitData", "type": "bytes"}
		],
		"name": "uninstallModule",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "mode", "type": "bytes32"},
			{"name": "executionCalldata", "type": "bytes"}
		],
		"name": "execute",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"},
			{"name": "module", "type": "address"},
			{"name": "additionalContext", "type": "bytes"}
		],
		"name": "isModuleInstalled",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "accountId",
		"outputs": [{"name": "", "type": "string"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "moduleTypeId", "type": "uint256"}
		],
		"name": "supportsModule",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// Safe7579Client provides typed access to the Safe7579 adapter
// contract.
type Safe7579Client struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewSafe7579Client creates a new Safe7579 client.
func NewSafe7579Client(
	caller contract.ContractCaller,
	address common.Address,
	chainID int64,
) *Safe7579Client {
	return &Safe7579Client{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: Safe7579ABI,
	}
}

// InstallModule installs an ERC-7579 module on the account.
func (c *Safe7579Client) InstallModule(
	ctx context.Context,
	moduleTypeID *big.Int,
	module common.Address,
	initData []byte,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "installModule",
			Args: []interface{}{
				moduleTypeID, module, initData,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"install module: %w", err,
		)
	}
	return result.TxHash, nil
}

// UninstallModule removes an ERC-7579 module from the account.
func (c *Safe7579Client) UninstallModule(
	ctx context.Context,
	moduleTypeID *big.Int,
	module common.Address,
	deInitData []byte,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "uninstallModule",
			Args: []interface{}{
				moduleTypeID, module, deInitData,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"uninstall module: %w", err,
		)
	}
	return result.TxHash, nil
}

// Execute executes calldata through the Safe7579 adapter.
func (c *Safe7579Client) Execute(
	ctx context.Context,
	mode [32]byte,
	executionCalldata []byte,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "execute",
			Args: []interface{}{
				mode, executionCalldata,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("execute: %w", err)
	}
	return result.TxHash, nil
}

// IsModuleInstalled checks if a module is installed on the account.
func (c *Safe7579Client) IsModuleInstalled(
	ctx context.Context,
	moduleTypeID *big.Int,
	module common.Address,
	additionalContext []byte,
) (bool, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "isModuleInstalled",
			Args: []interface{}{
				moduleTypeID, module, additionalContext,
			},
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"check module installed: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(bool); ok {
			return v, nil
		}
	}
	return false, nil
}

// AccountID returns the account's identifier string.
func (c *Safe7579Client) AccountID(
	ctx context.Context,
) (string, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "accountId",
			Args:    []interface{}{},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"get account id: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(string); ok {
			return v, nil
		}
	}
	return "", nil
}

// SupportsModule checks if the account supports a module type.
func (c *Safe7579Client) SupportsModule(
	ctx context.Context,
	moduleTypeID *big.Int,
) (bool, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "supportsModule",
			Args:    []interface{}{moduleTypeID},
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"check module support: %w", err,
		)
	}
	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(bool); ok {
			return v, nil
		}
	}
	return false, nil
}
