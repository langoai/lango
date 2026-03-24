package config

import "time"

// OrchestrationConfig configures the structured multi-agent control plane.
type OrchestrationConfig struct {
	// Mode selects the orchestration mode: "classic" (default) or "structured".
	// In structured mode, a CoordinatingExecutor wraps the agent executor
	// to apply delegation guard, budget policy, and recovery policy.
	Mode string `mapstructure:"mode" json:"mode"`

	// CircuitBreaker configures per-agent circuit breaking.
	CircuitBreaker CircuitBreakerCfg `mapstructure:"circuitBreaker" json:"circuitBreaker"`

	// Budget configures observational budget policy.
	Budget BudgetCfg `mapstructure:"budget" json:"budget"`

	// Recovery configures failure recovery policy.
	Recovery RecoveryCfg `mapstructure:"recovery" json:"recovery"`
}

// CircuitBreakerCfg configures per-agent circuit breaking.
type CircuitBreakerCfg struct {
	// FailureThreshold is consecutive failures before circuit opens (default: 3).
	FailureThreshold int `mapstructure:"failureThreshold" json:"failureThreshold"`

	// ResetTimeout is how long before half-open probe (default: 30s).
	ResetTimeout time.Duration `mapstructure:"resetTimeout" json:"resetTimeout"`
}

// BudgetCfg configures observational budget tracking.
type BudgetCfg struct {
	// ToolCallLimit mirrors the inner executor's max turns (default: 50).
	ToolCallLimit int `mapstructure:"toolCallLimit" json:"toolCallLimit"`

	// DelegationLimit is the max delegations before alerting (default: 15).
	DelegationLimit int `mapstructure:"delegationLimit" json:"delegationLimit"`

	// AlertThreshold is the percentage at which budget alerts fire (default: 0.8).
	AlertThreshold float64 `mapstructure:"alertThreshold" json:"alertThreshold"`
}

// RecoveryCfg configures failure recovery policy.
type RecoveryCfg struct {
	// MaxRetries is the maximum retry attempts on failure (default: 2).
	MaxRetries int `mapstructure:"maxRetries" json:"maxRetries"`

	// CircuitBreakerCooldown is the time before re-enabling a tripped agent (default: 5m).
	CircuitBreakerCooldown time.Duration `mapstructure:"circuitBreakerCooldown" json:"circuitBreakerCooldown"`
}

// TraceStoreConfig configures turn trace retention and cleanup.
type TraceStoreConfig struct {
	// MaxAge is the maximum age of traces before cleanup (default: 720h / 30 days).
	MaxAge time.Duration `mapstructure:"maxAge" json:"maxAge"`

	// MaxTraces is the maximum number of traces to retain (default: 10000).
	MaxTraces int `mapstructure:"maxTraces" json:"maxTraces"`

	// FailedTraceMultiplier extends retention for failed traces (default: 2).
	FailedTraceMultiplier int `mapstructure:"failedTraceMultiplier" json:"failedTraceMultiplier"`

	// CleanupInterval is how often the cleanup goroutine runs (default: 1h).
	CleanupInterval time.Duration `mapstructure:"cleanupInterval" json:"cleanupInterval"`
}

// OrchestrationDefaults returns an OrchestrationConfig with default values.
func OrchestrationDefaults() OrchestrationConfig {
	return OrchestrationConfig{
		Mode: "classic",
		CircuitBreaker: CircuitBreakerCfg{
			FailureThreshold: 3,
			ResetTimeout:     30 * time.Second,
		},
		Budget: BudgetCfg{
			ToolCallLimit:   50,
			DelegationLimit: 15,
			AlertThreshold:  0.8,
		},
		Recovery: RecoveryCfg{
			MaxRetries:             2,
			CircuitBreakerCooldown: 5 * time.Minute,
		},
	}
}

// TraceStoreDefaults returns a TraceStoreConfig with default values.
func TraceStoreDefaults() TraceStoreConfig {
	return TraceStoreConfig{
		MaxAge:                30 * 24 * time.Hour, // 30 days
		MaxTraces:             10000,
		FailedTraceMultiplier: 2,
		CleanupInterval:       time.Hour,
	}
}
