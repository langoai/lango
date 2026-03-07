package hub

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// VaultClient provides typed access to a LangoVault contract instance.
type VaultClient struct {
	caller  *contract.Caller
	address common.Address
	chainID int64
	abiJSON string
}

// NewVaultClient creates a vault client for a specific vault address.
func NewVaultClient(caller *contract.Caller, address common.Address, chainID int64) *VaultClient {
	return &VaultClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: vaultABIJSON,
	}
}

// Deposit deposits ERC-20 tokens into the vault.
func (c *VaultClient) Deposit(ctx context.Context) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "deposit",
	})
	if err != nil {
		return "", fmt.Errorf("vault deposit: %w", err)
	}
	return result.TxHash, nil
}

// SubmitWork submits work proof to the vault.
func (c *VaultClient) SubmitWork(ctx context.Context, workHash [32]byte) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "submitWork",
		Args:    []interface{}{workHash},
	})
	if err != nil {
		return "", fmt.Errorf("vault submit work: %w", err)
	}
	return result.TxHash, nil
}

// Release releases vault funds to the seller.
func (c *VaultClient) Release(ctx context.Context) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "release",
	})
	if err != nil {
		return "", fmt.Errorf("vault release: %w", err)
	}
	return result.TxHash, nil
}

// Refund refunds vault funds to the buyer.
func (c *VaultClient) Refund(ctx context.Context) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "refund",
	})
	if err != nil {
		return "", fmt.Errorf("vault refund: %w", err)
	}
	return result.TxHash, nil
}

// Dispute raises a dispute on the vault.
func (c *VaultClient) Dispute(ctx context.Context) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "dispute",
	})
	if err != nil {
		return "", fmt.Errorf("vault dispute: %w", err)
	}
	return result.TxHash, nil
}

// Resolve resolves a disputed vault.
func (c *VaultClient) Resolve(ctx context.Context, sellerFavor bool, sellerAmount, buyerAmount *big.Int) (string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "resolve",
		Args:    []interface{}{sellerFavor, sellerAmount, buyerAmount},
	})
	if err != nil {
		return "", fmt.Errorf("vault resolve: %w", err)
	}
	return result.TxHash, nil
}

// Status reads the vault's current status.
func (c *VaultClient) Status(ctx context.Context) (OnChainDealStatus, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "status",
	})
	if err != nil {
		return 0, fmt.Errorf("vault status: %w", err)
	}
	if len(result.Data) > 0 {
		if s, ok := result.Data[0].(uint8); ok {
			return OnChainDealStatus(s), nil
		}
	}
	return 0, fmt.Errorf("unexpected status result")
}

// Amount reads the vault's escrowed amount.
func (c *VaultClient) Amount(ctx context.Context) (*big.Int, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "amount",
	})
	if err != nil {
		return nil, fmt.Errorf("vault amount: %w", err)
	}
	if len(result.Data) > 0 {
		if a, ok := result.Data[0].(*big.Int); ok {
			return a, nil
		}
	}
	return nil, fmt.Errorf("unexpected amount result")
}
