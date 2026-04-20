package paymentapproval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateUpfrontPayment_ApproveLowRiskPrepay(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "2.0",
		Trust: TrustInput{
			Score: 0.95,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:       "10.00",
			RemainingBudget: "9.00",
			UserMaxPrepay:   "5.00",
			TransactionMode: "direct",
		},
	})

	assert.Equal(t, DecisionApprove, outcome.Decision)
	assert.Equal(t, ModePrepay, outcome.SuggestedMode)
	assert.Equal(t, AmountLow, outcome.AmountClass)
}

func TestEvaluateUpfrontPayment_RejectOnBudgetPolicyFailure(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "7.00",
		Trust: TrustInput{
			Score: 0.90,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:       "10.00",
			RemainingBudget: "9.00",
			UserMaxPrepay:   "5.00",
			TransactionMode: "direct",
		},
	})

	assert.Equal(t, DecisionReject, outcome.Decision)
	assert.Equal(t, ModeReject, outcome.SuggestedMode)
	assert.NotEmpty(t, outcome.PolicyCode)
}

func TestEvaluateUpfrontPayment_EscalateOnLowTrust(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "1.00",
		Trust: TrustInput{
			Score: 0.29,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:       "10.00",
			RemainingBudget: "9.00",
			UserMaxPrepay:   "5.00",
			TransactionMode: "direct",
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, ModeEscalate, outcome.SuggestedMode)
	assert.Equal(t, "escalate_low_trust", outcome.PolicyCode)
}

func TestEvaluateUpfrontPayment_EscalateOnHighAmountThreshold(t *testing.T) {
	below := EvaluateUpfrontPayment(Input{
		Amount: "99.99",
		Trust: TrustInput{
			Score: 0.80,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:       "500.00",
			RemainingBudget: "400.00",
			UserMaxPrepay:   "500.00",
			TransactionMode: "direct",
		},
	})

	assert.Equal(t, DecisionApprove, below.Decision)
	assert.Equal(t, AmountHigh, below.AmountClass)

	outcome := EvaluateUpfrontPayment(Input{
		Amount: "100.00",
		Trust: TrustInput{
			Score: 0.80,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:       "500.00",
			RemainingBudget: "400.00",
			UserMaxPrepay:   "500.00",
			TransactionMode: "direct",
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, ModeEscalate, outcome.SuggestedMode)
	assert.Equal(t, AmountCritical, outcome.AmountClass)
}
