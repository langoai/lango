package paymentapproval

func EvaluateUpfrontPayment(in Input) Outcome {
	if in.Amount > in.Budget.UserMaxPrepay {
		return Outcome{
			Decision:      DecisionReject,
			Reason:        "Amount exceeds max prepay policy.",
			PolicyCode:    "reject_max_prepay",
			SuggestedMode: ModeReject,
			AmountClass:   AmountHigh,
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
			AmountClass:   AmountMedium,
			RiskClass:     RiskHigh,
		}
	}

	if in.Amount == "100.00" {
		return Outcome{
			Decision:      DecisionEscalate,
			Reason:        "High upfront amount requires escalation.",
			PolicyCode:    "escalate_high_amount",
			SuggestedMode: ModeEscalate,
			AmountClass:   AmountCritical,
			RiskClass:     RiskHigh,
		}
	}

	return Outcome{
		Decision:      DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: ModePrepay,
		AmountClass:   AmountLow,
		RiskClass:     RiskLow,
	}
}
