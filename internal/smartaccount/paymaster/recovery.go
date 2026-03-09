package paymaster

import (
	"context"
	"fmt"
	"time"
)

// FallbackMode determines behavior when paymaster retries are exhausted.
type FallbackMode string

const (
	// FallbackAbort aborts the transaction when paymaster fails.
	FallbackAbort FallbackMode = "abort"
	// FallbackDirectGas falls back to direct gas payment (user pays gas).
	FallbackDirectGas FallbackMode = "direct"
)

// RecoveryConfig configures paymaster error recovery behavior.
type RecoveryConfig struct {
	MaxRetries   int
	BaseDelay    time.Duration
	FallbackMode FallbackMode
}

// DefaultRecoveryConfig returns sensible defaults.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		MaxRetries:   2,
		BaseDelay:    200 * time.Millisecond,
		FallbackMode: FallbackAbort,
	}
}

// RecoverableProvider wraps a PaymasterProvider with retry and fallback logic.
type RecoverableProvider struct {
	inner  PaymasterProvider
	config RecoveryConfig
}

// NewRecoverableProvider wraps a provider with recovery.
func NewRecoverableProvider(inner PaymasterProvider, cfg RecoveryConfig) *RecoverableProvider {
	return &RecoverableProvider{inner: inner, config: cfg}
}

// SponsorUserOp sponsors a UserOp with retry for transient errors.
func (p *RecoverableProvider) SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error) {
	var lastErr error
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		result, err := p.inner.SponsorUserOp(ctx, req)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Permanent errors: fail immediately.
		if IsPermanent(err) {
			return nil, err
		}

		// Non-transient unknown errors: fail immediately.
		if !IsTransient(err) {
			return nil, err
		}

		// Transient error: retry with exponential backoff.
		if attempt < p.config.MaxRetries {
			delay := p.config.BaseDelay * (1 << uint(attempt))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	// All retries exhausted.
	switch p.config.FallbackMode {
	case FallbackDirectGas:
		// Return empty paymasterAndData — the UserOp will use direct gas.
		return &SponsorResult{
			PaymasterAndData: []byte{},
		}, nil
	default:
		return nil, fmt.Errorf("paymaster retries exhausted: %w", lastErr)
	}
}

// Type returns the underlying provider type.
func (p *RecoverableProvider) Type() string {
	return p.inner.Type() + "+recovery"
}
