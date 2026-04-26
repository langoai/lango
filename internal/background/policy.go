package background

import "time"

const (
	defaultMaxRetryAttempts = 3
	defaultRetryBaseDelay   = 25 * time.Millisecond
)

// RetryPolicy captures the bounded automatic retry behavior for background work.
type RetryPolicy struct {
	MaxRetryAttempts int
	BaseDelay        time.Duration
}

// DefaultRetryPolicy preserves the existing bounded exponential backoff behavior.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetryAttempts: defaultMaxRetryAttempts,
		BaseDelay:        defaultRetryBaseDelay,
	}
}

func (p RetryPolicy) normalized() RetryPolicy {
	if p.MaxRetryAttempts <= 0 {
		p.MaxRetryAttempts = defaultMaxRetryAttempts
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = defaultRetryBaseDelay
	}
	return p
}

// ShouldScheduleRetry reports whether another automatic retry should be queued
// after the given failed attempt count.
func (p RetryPolicy) ShouldScheduleRetry(attemptCount int) bool {
	p = p.normalized()
	return attemptCount > 0 && attemptCount <= p.MaxRetryAttempts
}

// DelayForAttempt returns the exponential backoff delay for a failed attempt.
func (p RetryPolicy) DelayForAttempt(attemptCount int) time.Duration {
	p = p.normalized()
	if attemptCount <= 1 {
		return p.BaseDelay
	}

	delay := p.BaseDelay
	for i := 1; i < attemptCount; i++ {
		delay *= 2
	}
	return delay
}
