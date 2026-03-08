package smartaccount

import "errors"

var (
	ErrAccountNotDeployed     = errors.New("smart account not deployed")
	ErrSessionExpired         = errors.New("session key expired")
	ErrSessionRevoked         = errors.New("session key revoked")
	ErrPolicyViolation        = errors.New("session policy violation")
	ErrModuleNotInstalled     = errors.New("module not installed")
	ErrSpendLimitExceeded     = errors.New("spend limit exceeded")
	ErrInvalidSessionKey      = errors.New("invalid session key")
	ErrSessionNotFound        = errors.New("session key not found")
	ErrTargetNotAllowed       = errors.New("target address not allowed")
	ErrFunctionNotAllowed     = errors.New("function not allowed")
	ErrInvalidUserOp          = errors.New("invalid user operation")
	ErrBundlerError           = errors.New("bundler RPC error")
	ErrModuleAlreadyInstalled = errors.New("module already installed")
)

// PolicyViolationError provides details about why a policy check failed.
type PolicyViolationError struct {
	SessionID string
	Reason    string
}

func (e *PolicyViolationError) Error() string {
	return "policy violation for session " + e.SessionID + ": " + e.Reason
}

func (e *PolicyViolationError) Unwrap() error { return ErrPolicyViolation }
