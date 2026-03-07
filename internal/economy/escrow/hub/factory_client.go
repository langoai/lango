package hub

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/contract"
)

// FactoryClient provides typed access to the LangoVaultFactory contract.
type FactoryClient struct {
	caller  contract.ContractCaller
	address common.Address
	chainID int64
	abiJSON string
}

// NewFactoryClient creates a factory client for the given contract address.
func NewFactoryClient(caller contract.ContractCaller, address common.Address, chainID int64) *FactoryClient {
	return &FactoryClient{
		caller:  caller,
		address: address,
		chainID: chainID,
		abiJSON: factoryABIJSON,
	}
}

// CreateVault creates a new vault clone via the factory.
func (c *FactoryClient) CreateVault(ctx context.Context, seller, token common.Address, amount, deadline *big.Int, arbitrator common.Address) (*VaultInfo, string, error) {
	result, err := c.caller.Write(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "createVault",
		Args:    []interface{}{seller, token, amount, deadline, arbitrator},
	})
	if err != nil {
		return nil, "", fmt.Errorf("create vault: %w", err)
	}

	info := &VaultInfo{}
	if len(result.Data) >= 2 {
		if id, ok := result.Data[0].(*big.Int); ok {
			info.VaultID = id
		}
		if addr, ok := result.Data[1].(common.Address); ok {
			info.VaultAddress = addr
		}
	}
	return info, result.TxHash, nil
}

// GetVault returns the vault address for a given vault ID.
func (c *FactoryClient) GetVault(ctx context.Context, vaultID *big.Int) (common.Address, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "getVault",
		Args:    []interface{}{vaultID},
	})
	if err != nil {
		return common.Address{}, fmt.Errorf("get vault %s: %w", vaultID.String(), err)
	}
	if len(result.Data) > 0 {
		if addr, ok := result.Data[0].(common.Address); ok {
			return addr, nil
		}
	}
	return common.Address{}, fmt.Errorf("unexpected vault result")
}

// VaultCount returns the total number of vaults created.
func (c *FactoryClient) VaultCount(ctx context.Context) (*big.Int, error) {
	result, err := c.caller.Read(ctx, contract.ContractCallRequest{
		ChainID: c.chainID,
		Address: c.address,
		ABI:     c.abiJSON,
		Method:  "vaultCount",
	})
	if err != nil {
		return nil, fmt.Errorf("vault count: %w", err)
	}
	if len(result.Data) > 0 {
		if n, ok := result.Data[0].(*big.Int); ok {
			return n, nil
		}
	}
	return big.NewInt(0), nil
}
