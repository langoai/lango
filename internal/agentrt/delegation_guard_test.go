package agentrt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func TestDelegationGuard_CircuitBreaker(t *testing.T) {
	tests := []struct {
		give       string
		failures   int
		threshold  int
		wantState  CircuitState
		wantIsOpen bool
	}{
		{
			give:       "no failures — closed",
			failures:   0,
			threshold:  3,
			wantState:  CircuitClosed,
			wantIsOpen: false,
		},
		{
			give:       "below threshold — closed",
			failures:   2,
			threshold:  3,
			wantState:  CircuitClosed,
			wantIsOpen: false,
		},
		{
			give:       "at threshold — open",
			failures:   3,
			threshold:  3,
			wantState:  CircuitOpen,
			wantIsOpen: true,
		},
		{
			give:       "above threshold — open",
			failures:   5,
			threshold:  3,
			wantState:  CircuitOpen,
			wantIsOpen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			guard := NewDelegationGuard(config.CircuitBreakerCfg{
				FailureThreshold: tt.threshold,
				ResetTimeout:     1 * time.Hour,
			}, nil)

			for i := 0; i < tt.failures; i++ {
				guard.RecordOutcome("agent-a", false)
			}

			assert.Equal(t, tt.wantIsOpen, guard.IsOpen("agent-a"))
			assert.Equal(t, tt.wantState, guard.State("agent-a"))
		})
	}
}

func TestDelegationGuard_SuccessResetsFailures(t *testing.T) {
	guard := NewDelegationGuard(config.CircuitBreakerCfg{
		FailureThreshold: 3,
		ResetTimeout:     1 * time.Hour,
	}, nil)

	guard.RecordOutcome("agent-a", false)
	guard.RecordOutcome("agent-a", false)
	guard.RecordOutcome("agent-a", true) // reset
	guard.RecordOutcome("agent-a", false)

	assert.False(t, guard.IsOpen("agent-a"))
}

func TestDelegationGuard_HalfOpenAfterTimeout(t *testing.T) {
	guard := NewDelegationGuard(config.CircuitBreakerCfg{
		FailureThreshold: 1,
		ResetTimeout:     1 * time.Millisecond,
	}, nil)

	guard.RecordOutcome("agent-a", false)
	assert.True(t, guard.IsOpen("agent-a"))

	time.Sleep(5 * time.Millisecond)
	assert.False(t, guard.IsOpen("agent-a")) // half-open now
	assert.Equal(t, CircuitHalfOpen, guard.State("agent-a"))
}

func TestDelegationGuard_UnknownAgent(t *testing.T) {
	guard := NewDelegationGuard(config.CircuitBreakerCfg{
		FailureThreshold: 3,
		ResetTimeout:     1 * time.Hour,
	}, nil)

	assert.False(t, guard.IsOpen("unknown"))
	assert.Equal(t, CircuitClosed, guard.State("unknown"))
}

func TestDelegationGuard_ProviderFailureTracking(t *testing.T) {
	t.Run("provider circuit opens after threshold failures", func(t *testing.T) {
		guard := NewDelegationGuard(config.CircuitBreakerCfg{
			FailureThreshold: 3,
			ResetTimeout:     1 * time.Hour,
		}, nil)

		guard.RecordProviderFailure("openai", false)
		guard.RecordProviderFailure("openai", false)
		assert.False(t, guard.IsProviderOpen("openai"))

		guard.RecordProviderFailure("openai", false)
		assert.True(t, guard.IsProviderOpen("openai"))
	})

	t.Run("provider circuit independent of agent circuit", func(t *testing.T) {
		guard := NewDelegationGuard(config.CircuitBreakerCfg{
			FailureThreshold: 2,
			ResetTimeout:     1 * time.Hour,
		}, nil)

		// Open the provider circuit.
		guard.RecordProviderFailure("openai", false)
		guard.RecordProviderFailure("openai", false)
		assert.True(t, guard.IsProviderOpen("openai"))

		// Agent "operator" should not be affected.
		assert.False(t, guard.IsOpen("operator"))
		assert.Equal(t, CircuitClosed, guard.State("operator"))
	})

	t.Run("provider success resets provider circuit", func(t *testing.T) {
		guard := NewDelegationGuard(config.CircuitBreakerCfg{
			FailureThreshold: 2,
			ResetTimeout:     1 * time.Hour,
		}, nil)

		guard.RecordProviderFailure("anthropic", false)
		guard.RecordProviderFailure("anthropic", false)
		assert.True(t, guard.IsProviderOpen("anthropic"))

		// Half-open probe success closes it.
		guard.RecordProviderFailure("anthropic", true)
		assert.False(t, guard.IsProviderOpen("anthropic"))
	})

	t.Run("agent name does not collide with provider name", func(t *testing.T) {
		guard := NewDelegationGuard(config.CircuitBreakerCfg{
			FailureThreshold: 2,
			ResetTimeout:     1 * time.Hour,
		}, nil)

		// Record failures for agent named "openai".
		guard.RecordOutcome("openai", false)
		guard.RecordOutcome("openai", false)
		assert.True(t, guard.IsOpen("openai"))

		// Provider "openai" should not be open — uses different key.
		assert.False(t, guard.IsProviderOpen("openai"))
	})
}
