package paymaster

import "errors"

var (
	ErrPaymasterRejected      = errors.New("paymaster rejected sponsorship")
	ErrPaymasterTimeout       = errors.New("paymaster request timed out")
	ErrInsufficientToken      = errors.New("insufficient token balance for gas")
	ErrPaymasterNotConfigured = errors.New("paymaster not configured")
)

// IsTransient reports whether err is a transient paymaster error eligible for retry.
func IsTransient(err error) bool {
	return errors.Is(err, ErrPaymasterTimeout)
}

// IsPermanent reports whether err is a permanent paymaster error (no retry).
func IsPermanent(err error) bool {
	return errors.Is(err, ErrPaymasterRejected) || errors.Is(err, ErrInsufficientToken)
}
