package escrow

import (
	"context"
	"math/big"
)

// NoopSettler is a placeholder settlement executor that performs no on-chain operations.
type NoopSettler struct{}

var _ SettlementExecutor = (*NoopSettler)(nil)

func (NoopSettler) Lock(_ context.Context, _ string, _ *big.Int) error    { return nil }
func (NoopSettler) Release(_ context.Context, _ string, _ *big.Int) error { return nil }
func (NoopSettler) Refund(_ context.Context, _ string, _ *big.Int) error  { return nil }
