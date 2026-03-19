package runledger

import "errors"

// Sentinel errors for the RunLedger package.
var (
	ErrRunNotFound  = errors.New("run not found")
	ErrRunNotPaused = errors.New("run is not paused")
	ErrStepNotFound = errors.New("step not found")
	ErrAccessDenied = errors.New("access denied")
	ErrRunCompleted = errors.New("run already completed")
)
