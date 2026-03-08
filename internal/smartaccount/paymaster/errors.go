package paymaster

import "errors"

var (
	ErrPaymasterRejected     = errors.New("paymaster rejected sponsorship")
	ErrPaymasterTimeout      = errors.New("paymaster request timed out")
	ErrInsufficientToken     = errors.New("insufficient token balance for gas")
	ErrPaymasterNotConfigured = errors.New("paymaster not configured")
)
