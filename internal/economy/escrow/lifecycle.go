package escrow

import "fmt"

// validTransitions defines the allowed status transitions.
var validTransitions = map[EscrowStatus][]EscrowStatus{
	StatusPending:   {StatusFunded, StatusExpired},
	StatusFunded:    {StatusActive, StatusExpired},
	StatusActive:    {StatusCompleted, StatusDisputed, StatusExpired},
	StatusCompleted: {StatusReleased, StatusDisputed},
	StatusDisputed:  {StatusRefunded, StatusReleased},
	// Terminal states: StatusReleased, StatusExpired, StatusRefunded have no transitions.
}

// canTransition returns true if from -> to is a valid transition.
func canTransition(from, to EscrowStatus) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// validateTransition returns an error if the transition is invalid.
func validateTransition(from, to EscrowStatus) error {
	if !canTransition(from, to) {
		return fmt.Errorf("%q -> %q: %w", from, to, ErrInvalidTransition)
	}
	return nil
}
