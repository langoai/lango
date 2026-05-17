// Package workflow provides CLI commands for workflow management.
package workflow

import (
	"context"
	"fmt"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/workflow"
)

var cancelWorkflowRun = func(bootLoader func() (*bootstrap.Result, error), runID string) (string, error) {
	boot, err := bootLoader()
	if err != nil {
		return "", fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	engine := initEngine(boot)
	if engine == nil {
		return "", ErrWorkflowDisabled
	}

	if err := engine.Cancel(runID); err != nil {
		return "", fmt.Errorf("cancel workflow: %w", err)
	}

	return fmt.Sprintf("Workflow run %s cancelled.", runID), nil
}

var executeWorkflowDirect = func(boot *bootstrap.Result, w *workflow.Workflow) (*workflow.RunResult, error) {
	engine := initEngine(boot)
	if engine == nil {
		return nil, ErrWorkflowDisabled
	}
	return engine.Run(context.Background(), w)
}

// NewWorkflowCmd creates the workflow command with lazy bootstrap loading.
func NewWorkflowCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflow execution",
		Long:  "Run, monitor, and manage multi-step workflow pipelines defined in .flow.yaml files.",
	}

	cmd.AddCommand(newRunCmd(bootLoader))
	cmd.AddCommand(newWorkflowListCmd(bootLoader))
	cmd.AddCommand(newWorkflowStatusCmd(bootLoader))
	cmd.AddCommand(newWorkflowCancelCmd(bootLoader))
	cmd.AddCommand(newWorkflowHistoryCmd(bootLoader))
	cmd.AddCommand(newValidateCmd())

	return cmd
}

func newRunCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var schedule string

	cmd := &cobra.Command{
		Use:   "run <file.flow.yaml>",
		Short: "Run a workflow from a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Parse workflow YAML
			w, err := workflow.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("parse workflow %q: %w", filePath, err)
			}

			// Override schedule if provided
			if schedule != "" {
				w.Schedule = schedule
			}

			// Validate
			if err := workflow.Validate(w); err != nil {
				return fmt.Errorf("validate workflow: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Workflow: %s\n", w.Name)
			fmt.Fprintf(out, "Steps:    %d\n", len(w.Steps))
			if w.Schedule != "" {
				fmt.Fprintf(out, "Schedule: %s\n", w.Schedule)
			}

			// For direct execution (no schedule), we need the full app running.
			// The CLI can only validate and display — actual execution happens via the server.
			if w.Schedule != "" {
				fmt.Fprintln(out, "\nWorkflow has a schedule.")
				job, err := registerScheduledWorkflow(cmd.Context(), bootLoader, filePath, w)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Scheduled workflow registered as cron job (id: %s)\n", job.ID)
				fmt.Fprintf(out, "  Name: %s\n", job.Name)
				fmt.Fprintf(out, "  Schedule: %s %s\n", job.ScheduleType, job.Schedule)
				return nil
			}

			fmt.Fprintln(out, "\nWorkflow validated successfully.")
			fmt.Fprintln(out, "To execute, start the server with 'lango serve' and submit via API or TUI.")

			// If server is running, try to execute directly
			boot, err := bootLoader()
			if err != nil {
				fmt.Fprintln(out, "(Server not available for direct execution)")
				return nil
			}
			defer boot.Close()

			result, err := executeWorkflowDirect(boot, w)
			if err == ErrWorkflowDisabled {
				fmt.Fprintln(out, "(Workflow engine not enabled in config)")
				return nil
			}
			if err != nil {
				return fmt.Errorf("execute workflow: %w", err)
			}

			fmt.Fprintln(out, "\nExecuting workflow...")
			fmt.Fprintf(out, "\nWorkflow completed: %s\n", result.Status)
			if result.Error != "" {
				fmt.Fprintf(out, "Error: %s\n", result.Error)
			}
			for stepID, stepResult := range result.StepResults {
				fmt.Fprintf(out, "\n--- Step: %s ---\n%s\n", stepID, truncate(stepResult, 500))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&schedule, "schedule", "", "cron schedule to register (overrides YAML)")
	return cmd
}

func registerScheduledWorkflow(ctx context.Context, bootLoader func() (*bootstrap.Result, error), filePath string, w *workflow.Workflow) (*cron.Job, error) {
	boot, err := bootLoader()
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	defer boot.Close()

	store := boot.Storage.Cron()
	if store == nil {
		return nil, fmt.Errorf("workflow schedule registration unavailable: cron storage is not configured")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("resolve workflow path: %w", err)
	}

	job := cron.Job{
		ID:           uuid.New().String(),
		Name:         scheduledWorkflowJobName(w.Name),
		ScheduleType: "cron",
		Schedule:     w.Schedule,
		Prompt:       scheduledWorkflowPrompt(absPath),
		SessionMode:  "isolated",
		DeliverTo:    append([]string(nil), w.DeliverTo...),
		Timezone:     "UTC",
		Enabled:      true,
		CreatedAt:    time.Now(),
	}
	if err := cron.ValidateJobSchedule(job); err != nil {
		return nil, fmt.Errorf("invalid workflow schedule: %w", err)
	}
	if err := store.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("register scheduled workflow: %w", err)
	}
	return &job, nil
}

func scheduledWorkflowJobName(workflowName string) string {
	return "workflow:" + workflowName
}

func scheduledWorkflowPrompt(absPath string) string {
	return fmt.Sprintf("Run the saved workflow by calling the workflow_run tool with file_path %q. Report the workflow run id and final status.", absPath)
}

func newWorkflowListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			reader := workflowRunStore(boot, nil)
			if reader == nil {
				return ErrWorkflowDisabled
			}

			if runs, handled, err := maybeListRunsFromLedger(boot, limit); handled {
				if err != nil {
					return err
				}
				if len(runs) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No workflow runs found.")
					return nil
				}
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tSTEPS")
				for _, r := range runs {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\n",
						shortID(r.RunID), r.WorkflowName, r.Status, r.CompletedSteps, r.TotalSteps)
				}
				return w.Flush()
			}

			runs, err := reader.ListRuns(context.Background(), limit)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}

			if len(runs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workflow runs found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tSTEPS\tSTARTED")
			for _, r := range runs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\t%s\n",
					shortID(r.RunID), r.WorkflowName, r.Status,
					r.CompletedSteps, r.TotalSteps,
					formatTime(r.StartedAt))
			}
			return w.Flush()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "maximum entries to show")
	return cmd
}

func newWorkflowStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status <run-id>",
		Short: "Show workflow run status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			reader := workflowRunStore(boot, nil)
			if reader == nil {
				return ErrWorkflowDisabled
			}

			if status, handled, err := maybeStatusFromLedger(boot, args[0]); handled {
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "Run ID:    %s\n", status.RunID)
				fmt.Fprintf(out, "Workflow:  %s\n", status.WorkflowName)
				fmt.Fprintf(out, "Status:    %s\n", status.Status)
				fmt.Fprintf(out, "Progress:  %d/%d steps\n", status.CompletedSteps, status.TotalSteps)
				if len(status.StepStatuses) > 0 {
					fmt.Fprintln(out, "\nSteps:")
					for _, s := range status.StepStatuses {
						errInfo := ""
						if s.Error != "" {
							errInfo = " (" + truncate(s.Error, 40) + ")"
						}
						fmt.Fprintf(out, "  %-20s  %-12s  agent=%-15s%s\n",
							s.StepID, s.Status, s.Agent, errInfo)
					}
				}
				return nil
			}

			status, err := reader.GetRunStatus(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Run ID:    %s\n", status.RunID)
			fmt.Fprintf(out, "Workflow:  %s\n", status.WorkflowName)
			fmt.Fprintf(out, "Status:    %s\n", status.Status)
			fmt.Fprintf(out, "Progress:  %d/%d steps\n", status.CompletedSteps, status.TotalSteps)

			if len(status.StepStatuses) > 0 {
				fmt.Fprintln(out, "\nSteps:")
				for _, s := range status.StepStatuses {
					errInfo := ""
					if s.Error != "" {
						errInfo = " (" + truncate(s.Error, 40) + ")"
					}
					fmt.Fprintf(out, "  %-20s  %-12s  agent=%-15s%s\n",
						s.StepID, s.Status, s.Agent, errInfo)
				}
			}
			return nil
		},
	}
}

func newWorkflowCancelCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <run-id>",
		Short: "Cancel a running workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message, err := cancelWorkflowRun(bootLoader, args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), message)
			return nil
		},
	}
}

func newWorkflowHistoryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show workflow execution history",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			reader := workflowRunStore(boot, nil)
			if reader == nil {
				return ErrWorkflowDisabled
			}

			if runs, handled, err := maybeListRunsFromLedger(boot, limit); handled {
				if err != nil {
					return err
				}
				if len(runs) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No workflow history found.")
					return nil
				}
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tSTEPS")
				for _, r := range runs {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\n",
						shortID(r.RunID), r.WorkflowName, r.Status, r.CompletedSteps, r.TotalSteps)
				}
				return w.Flush()
			}

			runs, err := reader.ListRuns(context.Background(), limit)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}

			if len(runs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workflow history found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tSTEPS")
			for _, r := range runs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\n",
					shortID(r.RunID), r.WorkflowName, r.Status,
					r.CompletedSteps, r.TotalSteps)
			}
			return w.Flush()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "maximum entries to show")
	return cmd
}

func initEngine(boot *bootstrap.Result) *workflow.Engine {
	if !boot.Config.Workflow.Enabled {
		return nil
	}

	lg, _ := zap.NewProduction()
	state := workflowRunStore(boot, lg.Sugar())
	runStore, ok := state.(workflow.RunStore)
	if !ok || runStore == nil {
		return nil
	}
	return workflow.NewEngine(nil, runStore, nil,
		boot.Config.Workflow.MaxConcurrentSteps,
		boot.Config.Workflow.DefaultTimeout,
		lg.Sugar())
}

func maybeListRunsFromLedger(
	boot *bootstrap.Result,
	limit int,
) ([]workflow.RunStatus, bool, error) {
	if !boot.Config.RunLedger.Enabled || !boot.Config.RunLedger.AuthoritativeRead {
		return nil, false, nil
	}
	store := workflowRunLedgerStore(boot)
	if store == nil {
		return nil, true, fmt.Errorf("workflow runledger storage unavailable")
	}
	summaries, err := store.ListRuns(context.Background(), limit)
	if err != nil {
		return nil, true, fmt.Errorf("list workflow runs from RunLedger: %w", err)
	}
	result := make([]workflow.RunStatus, 0, len(summaries))
	for _, summary := range summaries {
		result = append(result, workflow.RunStatus{
			RunID:          summary.RunID,
			WorkflowName:   summary.Goal,
			Status:         string(summary.Status),
			TotalSteps:     summary.TotalSteps,
			CompletedSteps: summary.CompletedSteps,
		})
	}
	return result, true, nil
}

func maybeStatusFromLedger(
	boot *bootstrap.Result,
	runID string,
) (*workflow.RunStatus, bool, error) {
	if !boot.Config.RunLedger.Enabled || !boot.Config.RunLedger.AuthoritativeRead {
		return nil, false, nil
	}
	store := workflowRunLedgerStore(boot)
	if store == nil {
		return nil, true, fmt.Errorf("workflow runledger storage unavailable")
	}
	snap, err := store.GetRunSnapshot(context.Background(), runID)
	if err != nil {
		return nil, true, fmt.Errorf("get workflow status from RunLedger: %w", err)
	}
	statuses := make([]workflow.StepStatus, 0, len(snap.Steps))
	for _, step := range snap.Steps {
		statuses = append(statuses, workflow.StepStatus{
			StepID: step.StepID,
			Agent:  step.OwnerAgent,
			Status: string(step.Status),
			Error:  "",
		})
	}
	return &workflow.RunStatus{
		RunID:          snap.RunID,
		WorkflowName:   snap.Goal,
		Status:         string(snap.Status),
		TotalSteps:     len(snap.Steps),
		CompletedSteps: snap.CompletedSteps(),
		StartedAt:      snap.UpdatedAt,
		StepStatuses:   statuses,
	}, true, nil
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func truncate(s string, max int) string {
	return toolchain.Truncate(s, max)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.DateTime)
}

func workflowRunStore(boot *bootstrap.Result, logger *zap.SugaredLogger) storage.WorkflowRunReader {
	if boot != nil && boot.Storage != nil {
		return boot.Storage.WorkflowStateStore(logger)
	}
	return nil
}

func workflowRunLedgerStore(boot *bootstrap.Result) runledger.RunLedgerStore {
	if boot != nil && boot.Storage != nil {
		return boot.Storage.RunLedger()
	}
	return nil
}
