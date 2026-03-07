package app

import (
	"github.com/langoai/lango/internal/contract"
)

// contractComponents holds optional smart contract interaction components.
type contractComponents struct {
	caller *contract.Caller
}

// initContract creates the contract interaction components if payment is available.
func initContract(pc *paymentComponents) *contractComponents {
	if pc == nil {
		return nil
	}
	cache := contract.NewABICache()
	caller := contract.NewCaller(pc.rpcClient, pc.wallet, pc.chainID, cache)
	return &contractComponents{caller: caller}
}
