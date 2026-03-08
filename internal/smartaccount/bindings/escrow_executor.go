package bindings

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// EscrowExecutorABI is the ABI for LangoEscrowExecutor.
const EscrowExecutorABI = `[
	{
		"inputs": [
			{
				"components": [
					{"name": "target", "type": "address"},
					{"name": "value", "type": "uint256"},
					{"name": "callData", "type": "bytes"}
				],
				"name": "executions",
				"type": "tuple[]"
			}
		],
		"name": "executeBatchedEscrow",
		"outputs": [
			{"name": "results", "type": "bytes[]"}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "escrowId", "type": "bytes32"}
		],
		"name": "getEscrowStatus",
		"outputs": [
			{"name": "status", "type": "uint8"},
			{"name": "amount", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "escrowId", "type": "bytes32"}
		],
		"name": "releaseEscrow",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "escrowId", "type": "bytes32"}
		],
		"name": "refundEscrow",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// EscrowExecution represents a single execution in a batch.
type EscrowExecution struct {
	Target   common.Address
	Value    *big.Int
	CallData []byte
}

// EscrowExecutorClient provides typed access to the
// LangoEscrowExecutor contract.
type EscrowExecutorClient struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewEscrowExecutorClient creates a new escrow executor client.
func NewEscrowExecutorClient(
	caller contract.ContractCaller,
	address common.Address,
	chainID int64,
) *EscrowExecutorClient {
	return &EscrowExecutorClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: EscrowExecutorABI,
	}
}

// ExecuteBatchedEscrow executes a batch of escrow operations.
func (c *EscrowExecutorClient) ExecuteBatchedEscrow(
	ctx context.Context,
	executions []EscrowExecution,
) (string, error) {
	// Convert to ABI-compatible format.
	args := make([]interface{}, len(executions))
	for i, exec := range executions {
		value := exec.Value
		if value == nil {
			value = new(big.Int)
		}
		args[i] = struct {
			Target   common.Address
			Value    *big.Int
			CallData []byte
		}{
			Target:   exec.Target,
			Value:    value,
			CallData: exec.CallData,
		}
	}

	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "executeBatchedEscrow",
			Args:    []interface{}{args},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"execute batched escrow: %w", err,
		)
	}
	return result.TxHash, nil
}

// GetEscrowStatus returns the status and amount for an escrow.
func (c *EscrowExecutorClient) GetEscrowStatus(
	ctx context.Context,
	escrowID [32]byte,
) (uint8, *big.Int, error) {
	result, err := c.caller.Read(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "getEscrowStatus",
			Args:    []interface{}{escrowID},
		},
	)
	if err != nil {
		return 0, nil, fmt.Errorf(
			"get escrow status: %w", err,
		)
	}

	var status uint8
	amount := new(big.Int)

	if len(result.Data) > 0 {
		if v, ok := result.Data[0].(uint8); ok {
			status = v
		}
	}
	if len(result.Data) > 1 {
		if v, ok := result.Data[1].(*big.Int); ok {
			amount = v
		}
	}
	return status, amount, nil
}

// ReleaseEscrow releases funds from an escrow to the recipient.
func (c *EscrowExecutorClient) ReleaseEscrow(
	ctx context.Context,
	escrowID [32]byte,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "releaseEscrow",
			Args:    []interface{}{escrowID},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"release escrow: %w", err,
		)
	}
	return result.TxHash, nil
}

// RefundEscrow refunds funds from an escrow to the depositor.
func (c *EscrowExecutorClient) RefundEscrow(
	ctx context.Context,
	escrowID [32]byte,
) (string, error) {
	result, err := c.caller.Write(
		ctx, contract.ContractCallRequest{
			ChainID: c.chainID,
			Address: c.address,
			ABI:     c.abiJSON,
			Method:  "refundEscrow",
			Args:    []interface{}{escrowID},
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"refund escrow: %w", err,
		)
	}
	return result.TxHash, nil
}
