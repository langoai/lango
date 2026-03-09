package risk

import (
	"context"
	"math/big"
)

// Assessor evaluates transaction risk and recommends a payment strategy.
type Assessor interface {
	Assess(ctx context.Context, peerDID string, amount *big.Int, v Verifiability) (*Assessment, error)
}
