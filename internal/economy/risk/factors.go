package risk

import "math/big"

// computeFactors evaluates each dimension of the risk matrix.
func computeFactors(trust float64, amount *big.Int, threshold *big.Int, v Verifiability) []Factor {
	return []Factor{
		trustFactor(trust),
		amountFactor(amount, threshold),
		verifiabilityFactor(v),
	}
}

// trustFactor inverts trust: lower trust = higher risk.
func trustFactor(trust float64) Factor {
	return Factor{
		Name:   "trust",
		Value:  clamp(1.0 - trust),
		Weight: 0.4,
	}
}

// amountFactor normalizes transaction amount relative to the escrow threshold.
// Uses a sigmoid-like curve: risk = ratio / (1 + ratio).
func amountFactor(amount *big.Int, threshold *big.Int) Factor {
	var value float64
	if amount != nil && amount.Sign() > 0 && threshold != nil && threshold.Sign() > 0 {
		amountF := new(big.Float).SetInt(amount)
		threshF := new(big.Float).SetInt(threshold)
		ratio, _ := new(big.Float).Quo(amountF, threshF).Float64()
		value = ratio / (1.0 + ratio)
	}
	return Factor{
		Name:   "value",
		Value:  clamp(value),
		Weight: 0.35,
	}
}

// verifiabilityFactor maps verifiability level to risk.
func verifiabilityFactor(v Verifiability) Factor {
	var value float64
	switch v {
	case VerifiabilityHigh:
		value = 0.1
	case VerifiabilityMedium:
		value = 0.5
	case VerifiabilityLow:
		value = 0.9
	default:
		value = 0.9
	}
	return Factor{
		Name:   "verifiability",
		Value:  value,
		Weight: 0.25,
	}
}

// computeRiskScore calculates a weighted average of all factors.
func computeRiskScore(factors []Factor) float64 {
	var totalWeight, weightedSum float64
	for _, f := range factors {
		weightedSum += f.Value * f.Weight
		totalWeight += f.Weight
	}
	if totalWeight == 0 {
		return 0
	}
	return clamp(weightedSum / totalWeight)
}

// classifyRisk maps a continuous risk score to a discrete level.
func classifyRisk(score float64) RiskLevel {
	switch {
	case score < 0.3:
		return RiskLow
	case score < 0.6:
		return RiskMedium
	case score < 0.85:
		return RiskHigh
	default:
		return RiskCritical
	}
}
