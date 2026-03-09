package risk

import (
	"math/big"
	"time"
)

// RiskLevel represents the assessed risk of a transaction.
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// Strategy is the recommended payment strategy based on risk assessment.
type Strategy string

const (
	StrategyDirectPay    Strategy = "direct_pay"
	StrategyMicroPayment Strategy = "micro_payment"
	StrategyEscrow       Strategy = "escrow"
	StrategyZKFirst      Strategy = "zk_first"
	StrategyZKEscrow     Strategy = "zk_escrow"
)

// Verifiability describes how verifiable the work output is.
type Verifiability string

const (
	VerifiabilityHigh   Verifiability = "high"   // Output can be cryptographically verified
	VerifiabilityMedium Verifiability = "medium" // Output can be heuristically checked
	VerifiabilityLow    Verifiability = "low"    // Output requires manual review
)

// Factor represents a weighted risk factor.
type Factor struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`  // 0.0 to 1.0
	Weight float64 `json:"weight"` // relative weight
}

// Assessment is the result of a risk evaluation.
type Assessment struct {
	PeerDID       string        `json:"peerDid"`
	Amount        *big.Int      `json:"amount"`
	TrustScore    float64       `json:"trustScore"`
	Verifiability Verifiability `json:"verifiability"`
	RiskLevel     RiskLevel     `json:"riskLevel"`
	RiskScore     float64       `json:"riskScore"` // 0.0 (safe) to 1.0 (risky)
	Strategy      Strategy      `json:"strategy"`
	Factors       []Factor      `json:"factors"`
	Explanation   string        `json:"explanation"`
	AssessedAt    time.Time     `json:"assessedAt"`
}
