package policy

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// Validator pre-validates contract calls against policies.
type Validator struct{}

// NewValidator creates a new policy validator.
func NewValidator() *Validator { return &Validator{} }

// Check validates a ContractCall against a HarnessPolicy and spend tracker.
// Returns nil if the call is allowed.
func (v *Validator) Check(
	policy *HarnessPolicy,
	tracker *SpendTracker,
	call *sa.ContractCall,
) error {
	// Check max transaction amount.
	if policy.MaxTxAmount != nil && call.Value != nil {
		if call.Value.Cmp(policy.MaxTxAmount) > 0 {
			return fmt.Errorf(
				"value %s exceeds max %s: %w",
				call.Value, policy.MaxTxAmount, sa.ErrSpendLimitExceeded,
			)
		}
	}

	// Check allowed targets.
	if len(policy.AllowedTargets) > 0 {
		if !containsAddress(policy.AllowedTargets, call.Target) {
			return fmt.Errorf(
				"target %s: %w", call.Target.Hex(), sa.ErrTargetNotAllowed,
			)
		}
	}

	// Check allowed functions.
	if len(policy.AllowedFunctions) > 0 && call.FunctionSig != "" {
		if !containsString(policy.AllowedFunctions, call.FunctionSig) {
			return fmt.Errorf(
				"function %s: %w",
				call.FunctionSig, sa.ErrFunctionNotAllowed,
			)
		}
	}

	// Reset spend tracker windows if expired.
	if tracker != nil {
		tracker.ResetIfNeeded(time.Now())

		// Check daily limit.
		if policy.DailyLimit != nil && call.Value != nil {
			projected := new(big.Int).Add(tracker.DailySpent, call.Value)
			if projected.Cmp(policy.DailyLimit) > 0 {
				return fmt.Errorf(
					"daily spend %s + %s exceeds limit %s: %w",
					tracker.DailySpent, call.Value,
					policy.DailyLimit, sa.ErrSpendLimitExceeded,
				)
			}
		}

		// Check monthly limit.
		if policy.MonthlyLimit != nil && call.Value != nil {
			projected := new(big.Int).Add(tracker.MonthlySpent, call.Value)
			if projected.Cmp(policy.MonthlyLimit) > 0 {
				return fmt.Errorf(
					"monthly spend %s + %s exceeds limit %s: %w",
					tracker.MonthlySpent, call.Value,
					policy.MonthlyLimit, sa.ErrSpendLimitExceeded,
				)
			}
		}
	}

	return nil
}

// containsAddress checks if addr is in the slice.
func containsAddress(addrs []common.Address, addr common.Address) bool {
	for _, a := range addrs {
		if a == addr {
			return true
		}
	}
	return false
}

// containsString checks if s is in the slice.
func containsString(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}
