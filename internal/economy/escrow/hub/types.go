package hub

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// OnChainDealStatus represents the deal status on the smart contract.
type OnChainDealStatus uint8

const (
	DealStatusCreated       OnChainDealStatus = 0
	DealStatusDeposited     OnChainDealStatus = 1
	DealStatusWorkSubmitted OnChainDealStatus = 2
	DealStatusReleased      OnChainDealStatus = 3
	DealStatusRefunded      OnChainDealStatus = 4
	DealStatusDisputed      OnChainDealStatus = 5
	DealStatusResolved      OnChainDealStatus = 6
)

// String returns the human-readable status name.
func (s OnChainDealStatus) String() string {
	switch s {
	case DealStatusCreated:
		return "created"
	case DealStatusDeposited:
		return "deposited"
	case DealStatusWorkSubmitted:
		return "work_submitted"
	case DealStatusReleased:
		return "released"
	case DealStatusRefunded:
		return "refunded"
	case DealStatusDisputed:
		return "disputed"
	case DealStatusResolved:
		return "resolved"
	default:
		return "unknown"
	}
}

// OnChainDeal represents a deal as returned from the Hub contract.
type OnChainDeal struct {
	Buyer    common.Address
	Seller   common.Address
	Token    common.Address
	Amount   *big.Int
	Deadline *big.Int
	Status   OnChainDealStatus
	WorkHash [32]byte
}

// VaultInfo holds metadata about a vault created by the factory.
type VaultInfo struct {
	VaultID      *big.Int
	VaultAddress common.Address
	Buyer        common.Address
	Seller       common.Address
}
