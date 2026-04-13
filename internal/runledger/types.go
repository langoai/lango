package runledger

import (
	"encoding/json"
	"github.com/langoai/lango/internal/logging"
	"time"
)

// RunStatus is the lifecycle status of a Run.
type RunStatus string

const (
	RunStatusPlanning  RunStatus = "planning"
	RunStatusRunning   RunStatus = "running"
	RunStatusPaused    RunStatus = "paused"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

// StepStatus is the lifecycle status of a Step.
type StepStatus string

const (
	StepStatusPending       StepStatus = "pending"
	StepStatusInProgress    StepStatus = "in_progress"
	StepStatusVerifyPending StepStatus = "verify_pending"
	StepStatusCompleted     StepStatus = "completed"
	StepStatusFailed        StepStatus = "failed"
	StepStatusInterrupted   StepStatus = "interrupted"
)

// ValidatorType identifies a built-in validation strategy.
// Custom validators are intentionally not supported — no auto-pass allowed.
type ValidatorType string

const (
	ValidatorBuildPass            ValidatorType = "build_pass"
	ValidatorTestPass             ValidatorType = "test_pass"
	ValidatorFileChanged          ValidatorType = "file_changed"
	ValidatorArtifactExists       ValidatorType = "artifact_exists"
	ValidatorCommandPass          ValidatorType = "command_pass"
	ValidatorOrchestratorApproval ValidatorType = "orchestrator_approval"
)

// ValidatorSpec specifies how a step or acceptance criterion is validated.
type ValidatorSpec struct {
	Type    ValidatorType     `json:"type"`
	Target  string            `json:"target,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	WorkDir string            `json:"work_dir,omitempty"` // set at runtime by workspace manager
}

// Step represents a discrete unit of work within a Run.
type Step struct {
	StepID      string        `json:"step_id"`
	Index       int           `json:"index"`
	Goal        string        `json:"goal"`
	OwnerAgent  string        `json:"owner_agent"`
	Status      StepStatus    `json:"status"`
	Evidence    []Evidence    `json:"evidence,omitempty"`
	Validator   ValidatorSpec `json:"validator"`
	ToolProfile []string      `json:"tool_profile,omitempty"`
	RetryCount  int           `json:"retry_count"`
	MaxRetries  int           `json:"max_retries"`
	ResumeFrom  string        `json:"resume_from,omitempty"`
	Result      string        `json:"result,omitempty"`
	DependsOn   []string      `json:"depends_on,omitempty"`
}

// Evidence is a piece of proof attached to a step result proposal.
type Evidence struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// AcceptanceCriterion describes a requirement for Run completion.
type AcceptanceCriterion struct {
	Description string        `json:"description"`
	Validator   ValidatorSpec `json:"validator"`
	Met         bool          `json:"met"`
	MetAt       *time.Time    `json:"met_at,omitempty"`
}

// ValidationResult is the output of a Validator execution.
type ValidationResult struct {
	Passed  bool              `json:"passed"`
	Reason  string            `json:"reason"`
	Details map[string]string `json:"details,omitempty"`
	Missing []string          `json:"missing,omitempty"`
}

// PlannerOutput is the deserialized JSON returned by the Planner agent.
type PlannerOutput struct {
	Goal               string                    `json:"goal"`
	AcceptanceCriteria []AcceptanceCriteriaInput `json:"acceptance_criteria"`
	Steps              []StepInput               `json:"steps"`
}

// AcceptanceCriteriaInput is the planner's acceptance criterion format.
type AcceptanceCriteriaInput struct {
	Description string        `json:"description"`
	Validator   ValidatorSpec `json:"validator"`
}

// StepInput is the planner's step format.
type StepInput struct {
	ID          string        `json:"id"`
	Goal        string        `json:"goal"`
	OwnerAgent  string        `json:"owner_agent"`
	Validator   ValidatorSpec `json:"validator"`
	ToolProfile []string      `json:"tool_profile,omitempty"`
	DependsOn   []string      `json:"depends_on,omitempty"`
}

// RunSummary is a compressed representation of a Run for context injection.
type RunSummary struct {
	RunID              string    `json:"run_id"`
	Goal               string    `json:"goal"`
	Status             RunStatus `json:"status"`
	TotalSteps         int       `json:"total_steps"`
	CompletedSteps     int       `json:"completed_steps"`
	CurrentStepGoal    string    `json:"current_step_goal,omitempty"`
	CurrentStepStatus  string    `json:"current_step_status,omitempty"`
	CurrentBlocker     string    `json:"current_blocker,omitempty"`
	UnmetCriteria      []string  `json:"unmet_criteria,omitempty"`
	LastVerifiedResult string    `json:"last_verified_result,omitempty"`
}

// ToolProfile defines which tools are accessible during a step.
type ToolProfile string

const (
	ToolProfileCoding     ToolProfile = "coding"
	ToolProfileBrowser    ToolProfile = "browser"
	ToolProfileKnowledge  ToolProfile = "knowledge"
	ToolProfileSupervisor ToolProfile = "supervisor"
)

// DefaultMaxRetries is the default number of retries before escalation.
const DefaultMaxRetries = 2

// marshalPayload is a helper to serialize a typed payload into json.RawMessage.
func marshalPayload(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		logging.SubsystemSugar("runledger").Warnw("marshalPayload", "error", err)
		return json.RawMessage(`{}`)
	}
	return data
}
