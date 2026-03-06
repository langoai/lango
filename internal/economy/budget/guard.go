package budget

import "math/big"

// Guard enforces budget constraints for task spending.
type Guard interface {
	Check(taskID string, amount *big.Int) error
	Record(taskID string, entry SpendEntry) error
	Reserve(taskID string, amount *big.Int) (releaseFunc func(), err error)
}
