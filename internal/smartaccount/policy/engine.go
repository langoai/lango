package policy

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// RiskPolicyFunc generates policy constraints from risk assessment.
type RiskPolicyFunc func(
	ctx context.Context, peerDID string,
) (*HarnessPolicy, error)

// Engine manages policies per account.
type Engine struct {
	policies  map[common.Address]*HarnessPolicy
	trackers  map[common.Address]*SpendTracker
	riskFn    RiskPolicyFunc
	validator *Validator
	mu        sync.RWMutex
}

// New creates a new policy engine.
func New() *Engine {
	return &Engine{
		policies:  make(map[common.Address]*HarnessPolicy),
		trackers:  make(map[common.Address]*SpendTracker),
		validator: NewValidator(),
	}
}

// SetRiskPolicy sets the risk-driven policy generation callback.
func (e *Engine) SetRiskPolicy(fn RiskPolicyFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.riskFn = fn
}

// SetPolicy sets the harness policy for an account.
func (e *Engine) SetPolicy(account common.Address, policy *HarnessPolicy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[account] = policy
	// Initialize tracker if not present.
	if _, ok := e.trackers[account]; !ok {
		e.trackers[account] = NewSpendTracker()
	}
}

// GetPolicy returns the policy for an account.
func (e *Engine) GetPolicy(account common.Address) (*HarnessPolicy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	p, ok := e.policies[account]
	return p, ok
}

// Validate checks a call against the account's policy.
func (e *Engine) Validate(
	account common.Address, call *sa.ContractCall,
) error {
	e.mu.RLock()
	policy, ok := e.policies[account]
	if !ok {
		e.mu.RUnlock()
		return sa.ErrPolicyViolation
	}
	tracker := e.trackers[account]
	e.mu.RUnlock()

	return e.validator.Check(policy, tracker, call)
}

// RecordSpend records a successful spend against trackers.
func (e *Engine) RecordSpend(account common.Address, amount *big.Int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	tracker, ok := e.trackers[account]
	if !ok {
		tracker = NewSpendTracker()
		e.trackers[account] = tracker
	}
	tracker.DailySpent = new(big.Int).Add(tracker.DailySpent, amount)
	tracker.MonthlySpent = new(big.Int).Add(tracker.MonthlySpent, amount)
}

// MergePolicies merges master and task policies (intersection of permissions).
// The result uses the tighter constraint for each field.
func MergePolicies(master, task *HarnessPolicy) *HarnessPolicy {
	result := &HarnessPolicy{
		RequiredRiskScore: master.RequiredRiskScore,
	}

	// Use the higher risk score requirement.
	if task.RequiredRiskScore > master.RequiredRiskScore {
		result.RequiredRiskScore = task.RequiredRiskScore
	}

	// MaxTxAmount: use the smaller.
	result.MaxTxAmount = minBigInt(master.MaxTxAmount, task.MaxTxAmount)

	// DailyLimit: use the smaller.
	result.DailyLimit = minBigInt(master.DailyLimit, task.DailyLimit)

	// MonthlyLimit: use the smaller.
	result.MonthlyLimit = minBigInt(master.MonthlyLimit, task.MonthlyLimit)

	// AutoApproveBelow: use the smaller.
	result.AutoApproveBelow = minBigInt(
		master.AutoApproveBelow, task.AutoApproveBelow,
	)

	// AllowedTargets: intersection.
	if len(master.AllowedTargets) > 0 && len(task.AllowedTargets) > 0 {
		result.AllowedTargets = intersectAddresses(
			master.AllowedTargets, task.AllowedTargets,
		)
	} else if len(master.AllowedTargets) > 0 {
		result.AllowedTargets = copyAddresses(master.AllowedTargets)
	} else if len(task.AllowedTargets) > 0 {
		result.AllowedTargets = copyAddresses(task.AllowedTargets)
	}

	// AllowedFunctions: intersection.
	if len(master.AllowedFunctions) > 0 && len(task.AllowedFunctions) > 0 {
		result.AllowedFunctions = intersectStrings(
			master.AllowedFunctions, task.AllowedFunctions,
		)
	} else if len(master.AllowedFunctions) > 0 {
		result.AllowedFunctions = copyStrings(master.AllowedFunctions)
	} else if len(task.AllowedFunctions) > 0 {
		result.AllowedFunctions = copyStrings(task.AllowedFunctions)
	}

	return result
}

// minBigInt returns the smaller of a and b, handling nil values.
func minBigInt(a, b *big.Int) *big.Int {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return new(big.Int).Set(b)
	}
	if b == nil {
		return new(big.Int).Set(a)
	}
	if a.Cmp(b) < 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

// intersectAddresses returns addresses present in both slices.
func intersectAddresses(a, b []common.Address) []common.Address {
	set := make(map[common.Address]struct{}, len(a))
	for _, addr := range a {
		set[addr] = struct{}{}
	}
	var result []common.Address
	for _, addr := range b {
		if _, ok := set[addr]; ok {
			result = append(result, addr)
		}
	}
	return result
}

// intersectStrings returns strings present in both slices.
func intersectStrings(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	var result []string
	for _, s := range b {
		if _, ok := set[s]; ok {
			result = append(result, s)
		}
	}
	return result
}

// copyAddresses returns a copy of the address slice.
func copyAddresses(src []common.Address) []common.Address {
	dst := make([]common.Address, len(src))
	copy(dst, src)
	return dst
}

// copyStrings returns a copy of the string slice.
func copyStrings(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
