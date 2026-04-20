package paymentapproval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateUpfrontPayment_ApproveLowRiskPrepay(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "1.00",
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
	assert.NotEmpty(t, outcome.PolicyCode)
}

func TestEvaluateUpfrontPayment_EscalateOnHighAmount(t *testing.T) {
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
}
