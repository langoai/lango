package workflow

import "time"

// Workflow represents a declared multi-step workflow.
type Workflow struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Schedule    string   `yaml:"schedule"`    // optional cron expression
	DeliverTo   []string `yaml:"deliver_to"`  // optional result delivery targets
	Steps       []Step   `yaml:"steps"`
}

// Step represents a single unit of work in a workflow.
type Step struct {
	ID        string        `yaml:"id"`
	Agent     string        `yaml:"agent"`      // executor | researcher | planner | memory-manager
	Prompt    string        `yaml:"prompt"`      // Go template with {{step-id.result}}
	DependsOn []string      `yaml:"depends_on"`
	DeliverTo []string      `yaml:"deliver_to"`  // per-step delivery
	Timeout   time.Duration `yaml:"timeout"`
}

// RunResult holds the final result of a workflow execution.
type RunResult struct {
	RunID        string
	WorkflowName string
	Status       string
	StepResults  map[string]string // stepID -> result
	Error        string
	StartedAt    time.Time
	CompletedAt  time.Time
}

// RunStatus holds the current status of a workflow execution.
type RunStatus struct {
	RunID          string
	WorkflowName   string
	Status         string
	TotalSteps     int
	CompletedSteps int
	StartedAt      time.Time
	StepStatuses   []StepStatus
}

// StepStatus holds the current status of a single step.
type StepStatus struct {
	StepID string
	Agent  string
	Status string
	Error  string
}
