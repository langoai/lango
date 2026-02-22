package circuits

import (
	"github.com/consensys/gnark/frontend"
)

// BalanceRangeCircuit proves that a private balance is greater than or equal
// to a public threshold, without revealing the actual balance value.
//
// Constraint: Balance >= Threshold
type BalanceRangeCircuit struct {
	Threshold frontend.Variable `gnark:",public"`
	Balance   frontend.Variable `gnark:""`
}

// Define implements frontend.Circuit and constrains the balance range proof.
func (c *BalanceRangeCircuit) Define(api frontend.API) error {
	// Cmp returns 1 if Balance > Threshold, -1 if Balance < Threshold, 0 if equal.
	// We need Balance >= Threshold, so Cmp(Balance, Threshold) must be 0 or 1.
	// Equivalent: Balance - Threshold >= 0, which AssertIsLessOrEqual(Threshold, Balance) enforces.
	api.AssertIsLessOrEqual(c.Threshold, c.Balance)
	return nil
}
