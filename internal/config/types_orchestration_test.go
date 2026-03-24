package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrchestrationDefaults(t *testing.T) {
	cfg := OrchestrationDefaults()

	assert.Equal(t, "classic", cfg.Mode)
	assert.Equal(t, 3, cfg.CircuitBreaker.FailureThreshold)
	assert.Equal(t, 30*time.Second, cfg.CircuitBreaker.ResetTimeout)
	assert.Equal(t, 50, cfg.Budget.ToolCallLimit)
	assert.Equal(t, 15, cfg.Budget.DelegationLimit)
	assert.Equal(t, 0.8, cfg.Budget.AlertThreshold)
	assert.Equal(t, 2, cfg.Recovery.MaxRetries)
	assert.Equal(t, 5*time.Minute, cfg.Recovery.CircuitBreakerCooldown)
}

func TestTraceStoreDefaults(t *testing.T) {
	cfg := TraceStoreDefaults()

	assert.Equal(t, 30*24*time.Hour, cfg.MaxAge)
	assert.Equal(t, 10000, cfg.MaxTraces)
	assert.Equal(t, 2, cfg.FailedTraceMultiplier)
	assert.Equal(t, time.Hour, cfg.CleanupInterval)
}

func TestOrchestrationConfig_ZeroValue(t *testing.T) {
	var cfg OrchestrationConfig
	assert.Empty(t, cfg.Mode)
	assert.Zero(t, cfg.CircuitBreaker.FailureThreshold)
	assert.Zero(t, cfg.Budget.ToolCallLimit)
	assert.Zero(t, cfg.Recovery.MaxRetries)
}
