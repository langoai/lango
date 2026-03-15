package sentinel

import (
	"sync"
	"testing"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestSessionGuard_CriticalAlert_TriggersRevoke(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveSeverity AlertSeverity
		wantRevoke   bool
	}{
		{
			give:         "critical severity triggers revoke",
			giveSeverity: SeverityCritical,
			wantRevoke:   true,
		},
		{
			give:         "high severity triggers revoke",
			giveSeverity: SeverityHigh,
			wantRevoke:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			bus := eventbus.New()
			guard := NewSessionGuard(bus)

			var mu sync.Mutex
			revoked := false
			guard.SetRevokeFunc(func() error {
				mu.Lock()
				defer mu.Unlock()
				revoked = true
				return nil
			})

			guard.Start()

			bus.Publish(SentinelAlertEvent{
				Alert: Alert{
					Severity: tt.giveSeverity,
					Type:     "test_alert",
					Message:  "test",
				},
			})

			mu.Lock()
			defer mu.Unlock()
			assert.Equal(t, tt.wantRevoke, revoked)
		})
	}
}

func TestSessionGuard_MediumAlert_TriggersRestrict(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)

	var mu sync.Mutex
	var restrictFactor float64
	guard.SetRestrictFunc(func(factor float64) error {
		mu.Lock()
		defer mu.Unlock()
		restrictFactor = factor
		return nil
	})

	guard.Start()

	bus.Publish(SentinelAlertEvent{
		Alert: Alert{
			Severity: SeverityMedium,
			Type:     "unusual_timing",
			Message:  "potential wash trade",
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.InDelta(t, 0.5, restrictFactor, 0.001)
}

func TestSessionGuard_LowAlert_NoAction(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)

	var mu sync.Mutex
	revoked := false
	restricted := false

	guard.SetRevokeFunc(func() error {
		mu.Lock()
		defer mu.Unlock()
		revoked = true
		return nil
	})
	guard.SetRestrictFunc(func(factor float64) error {
		mu.Lock()
		defer mu.Unlock()
		restricted = true
		return nil
	})

	guard.Start()

	bus.Publish(SentinelAlertEvent{
		Alert: Alert{
			Severity: SeverityLow,
			Type:     "info",
			Message:  "low severity event",
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.False(t, revoked, "low alert should not trigger revoke")
	assert.False(t, restricted, "low alert should not trigger restrict")
}

func TestSessionGuard_NilCallbacks(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)
	guard.Start()

	// Should not panic when callbacks are nil.
	bus.Publish(SentinelAlertEvent{
		Alert: Alert{
			Severity: SeverityCritical,
			Type:     "test",
			Message:  "no callbacks set",
		},
	})
}

func TestSessionGuard_Start_Idempotent(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)

	var mu sync.Mutex
	revokeCount := 0
	guard.SetRevokeFunc(func() error {
		mu.Lock()
		defer mu.Unlock()
		revokeCount++
		return nil
	})

	// Start twice — should only subscribe once.
	guard.Start()
	guard.Start()

	bus.Publish(SentinelAlertEvent{
		Alert: Alert{
			Severity: SeverityCritical,
			Type:     "test",
			Message:  "idempotent check",
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, revokeCount, "idempotent start should not double-subscribe")
}

func TestSessionGuard_Stop_DisablesAlertHandling(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)

	var mu sync.Mutex
	revoked := false
	guard.SetRevokeFunc(func() error {
		mu.Lock()
		defer mu.Unlock()
		revoked = true
		return nil
	})

	guard.Start()
	guard.Stop()

	// Alert after Stop() should be ignored.
	bus.Publish(SentinelAlertEvent{
		Alert: Alert{
			Severity: SeverityCritical,
			Type:     "test_after_stop",
			Message:  "should be ignored",
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.False(t, revoked, "alert after Stop() should not trigger revoke")
}

func TestSessionGuard_WrongEventType_Ignored(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	guard := NewSessionGuard(bus)

	revoked := false
	guard.SetRevokeFunc(func() error {
		revoked = true
		return nil
	})

	guard.Start()

	// Publish a different event type on the same topic.
	bus.Publish(eventbus.BudgetAlertEvent{
		TaskID:    "task-1",
		Threshold: 0.8,
	})

	assert.False(t, revoked, "wrong event type should be ignored")
}
