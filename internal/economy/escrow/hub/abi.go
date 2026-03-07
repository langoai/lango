// Package hub provides typed Go clients for the Lango on-chain escrow contracts.
package hub

import (
	_ "embed"
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/langoai/lango/internal/contract"
)

//go:embed abi/LangoEscrowHub.abi.json
var hubABIJSON string

//go:embed abi/LangoVault.abi.json
var vaultABIJSON string

//go:embed abi/LangoVaultFactory.abi.json
var factoryABIJSON string

// HubABIJSON returns the raw ABI JSON for LangoEscrowHub.
func HubABIJSON() string { return hubABIJSON }

// VaultABIJSON returns the raw ABI JSON for LangoVault.
func VaultABIJSON() string { return vaultABIJSON }

// FactoryABIJSON returns the raw ABI JSON for LangoVaultFactory.
func FactoryABIJSON() string { return factoryABIJSON }

// ParseHubABI parses the embedded Hub ABI.
func ParseHubABI() (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(hubABIJSON)
	if err != nil {
		return nil, fmt.Errorf("parse hub ABI: %w", err)
	}
	return parsed, nil
}

// ParseVaultABI parses the embedded Vault ABI.
func ParseVaultABI() (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(vaultABIJSON)
	if err != nil {
		return nil, fmt.Errorf("parse vault ABI: %w", err)
	}
	return parsed, nil
}

// ParseFactoryABI parses the embedded Factory ABI.
func ParseFactoryABI() (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(factoryABIJSON)
	if err != nil {
		return nil, fmt.Errorf("parse factory ABI: %w", err)
	}
	return parsed, nil
}
