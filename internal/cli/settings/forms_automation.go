package settings

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewCronForm creates the Cron Scheduler configuration form.
func NewCronForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Cron Scheduler Configuration")

	form.AddField(&tuicore.Field{
		Key: "cron_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Cron.Enabled,
		Description: "Enable the cron scheduler for recurring automated tasks",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_timezone", Label: "Timezone", Type: tuicore.InputText,
		Value:       cfg.Cron.Timezone,
		Placeholder: "UTC or Asia/Seoul",
		Description: "IANA timezone for cron schedule evaluation (e.g. America/New_York)",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_max_jobs", Label: "Max Concurrent Jobs", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Cron.MaxConcurrentJobs),
		Description: "Maximum number of cron jobs that can run simultaneously",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	sessionMode := cfg.Cron.DefaultSessionMode
	if sessionMode == "" {
		sessionMode = "isolated"
	}
	form.AddField(&tuicore.Field{
		Key: "cron_session_mode", Label: "Session Mode", Type: tuicore.InputSelect,
		Value:       sessionMode,
		Options:     []string{"isolated", "main"},
		Description: "isolated=separate session per job, main=shared with main conversation",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_history_retention", Label: "History Retention", Type: tuicore.InputText,
		Value:       cfg.Cron.HistoryRetention,
		Placeholder: "30d or 720h",
		Description: "How long to keep cron job execution history",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Cron.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver cron job results to",
	})

	return &form
}

// NewBackgroundForm creates the Background Tasks configuration form.
func NewBackgroundForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Background Tasks Configuration")

	form.AddField(&tuicore.Field{
		Key: "bg_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Background.Enabled,
		Description: "Enable asynchronous background task execution",
	})

	form.AddField(&tuicore.Field{
		Key: "bg_yield_ms", Label: "Yield Time (ms)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Background.YieldMs),
		Description: "Milliseconds to yield between task steps to avoid CPU monopolization",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "bg_max_tasks", Label: "Max Concurrent Tasks", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Background.MaxConcurrentTasks),
		Description: "Maximum number of background tasks running at the same time",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "bg_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Background.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver background task results to",
	})

	return &form
}

// NewWorkflowForm creates the Workflow Engine configuration form.
func NewWorkflowForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Workflow Engine Configuration")

	form.AddField(&tuicore.Field{
		Key: "wf_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Workflow.Enabled,
		Description: "Enable the DAG-based workflow engine for multi-step pipelines",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_max_steps", Label: "Max Concurrent Steps", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Workflow.MaxConcurrentSteps),
		Description: "Maximum workflow steps executed in parallel",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "wf_timeout", Label: "Default Timeout", Type: tuicore.InputText,
		Value:       cfg.Workflow.DefaultTimeout.String(),
		Placeholder: "10m",
		Description: "Default timeout for an entire workflow execution",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_state_dir", Label: "State Directory", Type: tuicore.InputText,
		Value:       cfg.Workflow.StateDir,
		Placeholder: "~/.lango/workflows",
		Description: "Directory to persist workflow state and checkpoints",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Workflow.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver workflow completion results to",
	})

	return &form
}

// NewRunLedgerForm creates the RunLedger configuration form.
func NewRunLedgerForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("RunLedger Configuration")

	form.AddField(&tuicore.Field{
		Key: "runledger_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.RunLedger.Enabled,
		Description: "Enable the RunLedger durable execution engine",
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_shadow", Label: "Shadow Mode", Type: tuicore.InputBool,
		Checked:     cfg.RunLedger.Shadow,
		Description: "Record journal events without making RunLedger authoritative",
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_write_through", Label: "Write-Through", Type: tuicore.InputBool,
		Checked:     cfg.RunLedger.WriteThrough,
		Description: "Mirror creates and updates through RunLedger before projections",
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_authoritative_read", Label: "Authoritative Read", Type: tuicore.InputBool,
		Checked:     cfg.RunLedger.AuthoritativeRead,
		Description: "Read run state from RunLedger snapshots instead of legacy projections",
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_workspace_isolation", Label: "Workspace Isolation", Type: tuicore.InputBool,
		Checked:     cfg.RunLedger.WorkspaceIsolation,
		Description: "Enable runtime worktree isolation for coding-step validation",
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_stale_ttl", Label: "Stale TTL", Type: tuicore.InputText,
		Value:       cfg.RunLedger.StaleTTL.String(),
		Placeholder: "1h",
		Description: "How long a paused run remains resumable",
		Validate: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("must be a valid duration")
			}
			if d <= 0 {
				return fmt.Errorf("must be greater than 0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_max_history", Label: "Max Run History", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.RunLedger.MaxRunHistory),
		Description: "Maximum number of runs to keep (0 = unlimited)",
		Validate: func(s string) error {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be an integer")
			}
			if i < 0 {
				return fmt.Errorf("must be 0 or greater")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_validator_timeout", Label: "Validator Timeout", Type: tuicore.InputText,
		Value:       cfg.RunLedger.ValidatorTimeout.String(),
		Placeholder: "2m",
		Description: "Timeout for individual validator execution",
		Validate: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("must be a valid duration")
			}
			if d <= 0 {
				return fmt.Errorf("must be greater than 0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "runledger_planner_retries", Label: "Planner Max Retries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.RunLedger.PlannerMaxRetries),
		Description: "Maximum retries for malformed planner output handling",
		Validate: func(s string) error {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be an integer")
			}
			if i < 0 {
				return fmt.Errorf("must be 0 or greater")
			}
			return nil
		},
	})

	return &form
}

// NewProvenanceForm creates the Provenance configuration form.
func NewProvenanceForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Provenance Configuration")

	form.AddField(&tuicore.Field{
		Key: "provenance_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Provenance.Enabled,
		Description: "Enable the session provenance subsystem",
	})

	form.AddField(&tuicore.Field{
		Key: "provenance_auto_on_step_complete", Label: "Auto on Step Complete", Type: tuicore.InputBool,
		Checked:     cfg.Provenance.Checkpoints.AutoOnStepComplete,
		Description: "Automatically create a checkpoint when a RunLedger step validates successfully",
	})

	form.AddField(&tuicore.Field{
		Key: "provenance_auto_on_policy", Label: "Auto on Policy", Type: tuicore.InputBool,
		Checked:     cfg.Provenance.Checkpoints.AutoOnPolicy,
		Description: "Automatically create a checkpoint when a RunLedger policy decision is applied",
	})

	form.AddField(&tuicore.Field{
		Key: "provenance_max_per_session", Label: "Max Per Session", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Provenance.Checkpoints.MaxPerSession),
		Description: "Maximum checkpoints to keep per session (0 = unlimited)",
		Validate: func(s string) error {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be an integer")
			}
			if i < 0 {
				return fmt.Errorf("must be 0 or greater")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "provenance_retention_days", Label: "Retention Days", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Provenance.Checkpoints.RetentionDays),
		Description: "How long checkpoints are retained before pruning (0 = unlimited)",
		Validate: func(s string) error {
			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be an integer")
			}
			if i < 0 {
				return fmt.Errorf("must be 0 or greater")
			}
			return nil
		},
	})

	return &form
}
