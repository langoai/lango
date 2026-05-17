package paymentapproval

import (
	"math/big"

	"github.com/langoai/lango/internal/finance"
)

var (
	lowAmountThreshold    = big.NewInt(5_000_000)
	mediumAmountThreshold = big.NewInt(50_000_000)
	highAmountThreshold   = big.NewInt(100_000_000)
)

func EvaluateUpfrontPayment(in Input) Outcome {
	amount, err := finance.ParseUSDC(in.Amount)
	if err != nil {
		return invalidAmountOutcome("invalid_amount", "Invalid upfront payment amount.")
	}

	userMaxPrepay, err := finance.ParseUSDC(in.Budget.UserMaxPrepay)
	if err != nil {
		return invalidAmountOutcome("invalid_user_max_prepay", "Invalid user max prepay policy.")
	}

	if amount.Cmp(userMaxPrepay) > 0 {
		return Outcome{
			Decision:      DecisionReject,
			Reason:        "Amount exceeds max prepay policy.",
			PolicyCode:    "reject_max_prepay",
			SuggestedMode: ModeReject,
			AmountClass:   classifyAmount(amount),
			RiskClass:     RiskMedium,
			FailureDetail: "user_max_prepay_exceeded",
		}
	}

	if in.Trust.Score < 0.3 {
		return Outcome{
			Decision:      DecisionEscalate,
			Reason:        "Trust score is in an edge-case range for upfront payment.",
			PolicyCode:    "escalate_low_trust",
			SuggestedMode: ModeEscalate,
			AmountClass:   classifyAmount(amount),
			RiskClass:     RiskHigh,
		}
	}

	if amount.Cmp(highAmountThreshold) >= 0 {
		return Outcome{
			Decision:      DecisionEscalate,
			Reason:        "High upfront amount requires escalation.",
			PolicyCode:    "escalate_high_amount",
			SuggestedMode: ModeEscalate,
			AmountClass:   classifyAmount(amount),
			RiskClass:     RiskHigh,
		}
	}

	return Outcome{
		Decision:      DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: ModePrepay,
		AmountClass:   classifyAmount(amount),
		RiskClass:     RiskLow,
	}
}

func classifyAmount(amount *big.Int) AmountClass {
	switch {
	case amount.Cmp(lowAmountThreshold) < 0:
		return AmountLow
	case amount.Cmp(mediumAmountThreshold) < 0:
		return AmountMedium
	case amount.Cmp(highAmountThreshold) < 0:
		return AmountHigh
	default:
		return AmountCritical
	}
}

func invalidAmountOutcome(code, reason string) Outcome {
	return Outcome{
		Decision:      DecisionReject,
		Reason:        reason,
		PolicyCode:    code,
		SuggestedMode: ModeReject,
		RiskClass:     RiskHigh,
		FailureDetail: code,
	}
}
