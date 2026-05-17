package paymentapproval

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestEvaluateUpfrontPayment_RejectsInvalidRuntimeAmounts(t *testing.T) {
	t.Parallel()

	t.Run("invalid payment amount", func(t *testing.T) {
		outcome := EvaluateUpfrontPayment(Input{
			Amount: "not-usdc",
			Trust:  TrustInput{Score: 0.90},
			Budget: BudgetPolicyContext{
				UserMaxPrepay: "5.00",
			},
		})

		assert.Equal(t, DecisionReject, outcome.Decision)
		assert.Equal(t, ModeReject, outcome.SuggestedMode)
		assert.Equal(t, "invalid_amount", outcome.PolicyCode)
		assert.Equal(t, "invalid_amount", outcome.FailureDetail)
	})

	t.Run("invalid user max prepay", func(t *testing.T) {
		outcome := EvaluateUpfrontPayment(Input{
			Amount: "1.00",
			Trust:  TrustInput{Score: 0.90},
			Budget: BudgetPolicyContext{
				UserMaxPrepay: "not-usdc",
			},
		})

		assert.Equal(t, DecisionReject, outcome.Decision)
		assert.Equal(t, ModeReject, outcome.SuggestedMode)
		assert.Equal(t, "invalid_user_max_prepay", outcome.PolicyCode)
		assert.Equal(t, "invalid_user_max_prepay", outcome.FailureDetail)
	})
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

func TestPaymentApprovalProductionCodeDoesNotPanic(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	var offenders []string
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		require.NoError(t, err)

		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := call.Fun.(*ast.Ident)
			if ok && ident.Name == "panic" {
				offenders = append(offenders, fset.Position(call.Pos()).String())
			}
			return true
		})
	}

	assert.Empty(t, offenders, "production paymentapproval code must return outcomes or errors instead of panicking")
}
