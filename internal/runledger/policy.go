package runledger

// PolicyAction identifies the orchestrator's response to a step failure.
type PolicyAction string

const (
	PolicyRetry           PolicyAction = "retry"
	PolicyDecompose       PolicyAction = "decompose"
	PolicyChangeAgent     PolicyAction = "change_agent"
	PolicyChangeValidator PolicyAction = "change_validator"
	PolicySkip            PolicyAction = "skip"
	PolicyAbort           PolicyAction = "abort"
	PolicyEscalate        PolicyAction = "escalate"
)

// PolicyRequest is generated when a step fails validation.
// The orchestrator must respond with a PolicyDecision.
type PolicyRequest struct {
	RunID      string            `json:"run_id"`
	StepID     string            `json:"step_id"`
	StepGoal   string            `json:"step_goal"`
	Failure    *ValidationResult `json:"failure"`
	RetryCount int               `json:"retry_count"`
	MaxRetries int               `json:"max_retries"`
}

// PolicyDecision is the orchestrator's action in response to a PolicyRequest.
type PolicyDecision struct {
	Action       PolicyAction   `json:"action"`
	NewSteps     []Step         `json:"new_steps,omitempty"`
	NewValidator *ValidatorSpec `json:"new_validator,omitempty"`
	NewAgent     string         `json:"new_agent,omitempty"`
	Reason       string         `json:"reason,omitempty"`
}
