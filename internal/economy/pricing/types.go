package pricing

import (
	"math/big"
	"time"
)

// PriceModifierType identifies the type of price modification.
type PriceModifierType string

const (
	ModifierTrustDiscount  PriceModifierType = "trust_discount"
	ModifierVolumeDiscount PriceModifierType = "volume_discount"
	ModifierSurge          PriceModifierType = "surge"
	ModifierCustom         PriceModifierType = "custom"
)

// PriceModifier represents an adjustment to a base price.
type PriceModifier struct {
	Type        PriceModifierType `json:"type"`
	Description string            `json:"description"`
	Factor      float64           `json:"factor"` // multiplier: 0.9 = 10% discount, 1.2 = 20% surge
}

// Quote represents a computed price for a tool invocation.
type Quote struct {
	ToolName   string          `json:"toolName"`
	BasePrice  *big.Int        `json:"basePrice"` // in smallest USDC units
	FinalPrice *big.Int        `json:"finalPrice"`
	Currency   string          `json:"currency"` // "USDC"
	Modifiers  []PriceModifier `json:"modifiers"`
	IsFree     bool            `json:"isFree"`
	ValidUntil time.Time       `json:"validUntil"`
	PeerDID    string          `json:"peerDid,omitempty"`
}
