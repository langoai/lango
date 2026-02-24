package security

import (
	"context"
	"time"
)

// withRetry retries op with exponential backoff for transient KMS errors.
// Base delay is 100ms, doubled each attempt. Only errors where IsTransient
// returns true are retried.
func withRetry(ctx context.Context, maxRetries int, op func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = op()
		if lastErr == nil {
			return nil
		}
		if !IsTransient(lastErr) {
			return lastErr
		}
		if attempt < maxRetries {
			delay := 100 * time.Millisecond * (1 << uint(attempt))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}
