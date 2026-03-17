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

//go:embed abi/LangoEscrowHubV2.abi.json
var hubV2ABIJSON string

//go:embed abi/LangoVaultV2.abi.json
var vaultV2ABIJSON string

// HubABIJSON returns the raw ABI JSON for LangoEscrowHub.
func HubABIJSON() string { return hubABIJSON }

// VaultABIJSON returns the raw ABI JSON for LangoVault.
func VaultABIJSON() string { return vaultABIJSON }

// FactoryABIJSON returns the raw ABI JSON for LangoVaultFactory.
func FactoryABIJSON() string { return factoryABIJSON }

// HubV2ABIJSON returns the raw ABI JSON for LangoEscrowHubV2.
func HubV2ABIJSON() string { return hubV2ABIJSON }

// VaultV2ABIJSON returns the raw ABI JSON for LangoVaultV2.
func VaultV2ABIJSON() string { return vaultV2ABIJSON }

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

// ParseHubV2ABI parses the embedded Hub V2 ABI.
func ParseHubV2ABI() (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(hubV2ABIJSON)
	if err != nil {
		return nil, fmt.Errorf("parse hub v2 ABI: %w", err)
	}
	return parsed, nil
}

// ParseVaultV2ABI parses the embedded Vault V2 ABI.
func ParseVaultV2ABI() (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(vaultV2ABIJSON)
	if err != nil {
		return nil, fmt.Errorf("parse vault v2 ABI: %w", err)
	}
	return parsed, nil
}
