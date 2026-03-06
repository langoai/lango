// Package contract provides generic smart contract interaction for EVM chains.
package contract

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ContractCallRequest holds parameters for a contract call.
type ContractCallRequest struct {
	ChainID int64          `json:"chainId"`
	Address common.Address `json:"address"`
	ABI     string         `json:"abi"`     // JSON ABI string
	Method  string         `json:"method"`
	Args    []interface{}  `json:"args"`
	Value   *big.Int       `json:"value,omitempty"` // ETH value for payable functions
}

// ContractCallResult holds the result of a contract call.
type ContractCallResult struct {
	Data    []interface{} `json:"data"`
	TxHash  string        `json:"txHash,omitempty"`
	GasUsed uint64        `json:"gasUsed,omitempty"`
}

// ParseABI parses a JSON ABI string into a go-ethereum ABI object.
func ParseABI(abiJSON string) (*abi.ABI, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
