package budget

// Option configures the Engine.
type Option func(*Engine)

// WithAlertCallback sets the callback invoked when budget crosses a threshold.
// The callback receives the taskID and the threshold percentage that was crossed.
func WithAlertCallback(fn func(taskID string, pct float64)) Option {
	return func(e *Engine) { e.alertCallback = fn }
}

// WithRiskAssessor sets the risk assessor used during Check.
func WithRiskAssessor(fn RiskAssessor) Option {
	return func(e *Engine) { e.riskAssessor = fn }
}
