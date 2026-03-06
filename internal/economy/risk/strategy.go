package risk

import "math/big"

// selectStrategy uses the 3-variable matrix (trust x value x verifiability)
// to pick a payment strategy.
//
// Matrix logic:
//
//	Amount > escrowThreshold (forced):
//	  High trust    → Escrow
//	  Medium trust  → Escrow
//	  Low trust     → ZKEscrow
//
//	Amount <= escrowThreshold:
//	  High trust + any verifiability                → DirectPay
//	  Medium trust + high verifiability             → DirectPay
//	  Medium trust + medium verifiability           → MicroPayment
//	  Medium trust + low verifiability              → Escrow
//	  Low trust + high verifiability                → MicroPayment
//	  Low trust + medium/low verifiability          → ZKFirst
func (e *Engine) selectStrategy(trust float64, amount *big.Int, v Verifiability) Strategy {
	highValue := amount.Cmp(e.escrowThreshold) > 0

	// High-value transactions force escrow-based strategies.
	if highValue {
		if trust < e.medTrust {
			return StrategyZKEscrow
		}
		return StrategyEscrow
	}

	switch {
	// High trust peer
	case trust >= e.highTrust:
		return StrategyDirectPay

	// Medium trust peer
	case trust >= e.medTrust:
		switch v {
		case VerifiabilityHigh:
			return StrategyDirectPay
		case VerifiabilityMedium:
			return StrategyMicroPayment
		default:
			return StrategyEscrow
		}

	// Low trust peer
	default:
		if v == VerifiabilityHigh {
			return StrategyMicroPayment
		}
		return StrategyZKFirst
	}
}
