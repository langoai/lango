// Package bindings provides Go ABI bindings for smart account contracts.
package bindings

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/langoai/lango/internal/contract"
)

// ParseABI parses a JSON ABI string.
func ParseABI(abiJSON string) (*ethabi.ABI, error) {
	parsed, err := contract.ParseABI(abiJSON)
	if err != nil {
		return nil, fmt.Errorf("parse ABI: %w", err)
	}
	return parsed, nil
}
