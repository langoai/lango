// Package workflow provides CLI commands for workflow management.
package workflow

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/workflow"
)

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

			fmt.Printf("Workflow: %s\n", w.Name)
			fmt.Printf("Steps:    %d\n", len(w.Steps))
			if w.Schedule != "" {
				fmt.Printf("Schedule: %s\n", w.Schedule)
			}

			// For direct execution (no schedule), we need the full app running.
			// The CLI can only validate and display — actual execution happens via the server.
			if w.Schedule != "" {
				fmt.Println("\nWorkflow has a schedule. Register it with the running server:")
				fmt.Printf("  POST /api/workflow/register with the YAML content\n")
				return nil
			}

			fmt.Println("\nWorkflow validated successfully.")
			fmt.Println("To execute, start the server with 'lango serve' and submit via API or TUI.")

			// If server is running, try to execute directly
			boot, err := bootLoader()
			if err != nil {
				fmt.Println("(Server not available for direct execution)")
				return nil
			}
			defer boot.DBClient.Close()

			engine := initEngine(boot)
			if engine == nil {
				fmt.Println("(Workflow engine not enabled in config)")
				return nil
			}

			fmt.Println("\nExecuting workflow...")
			result, err := engine.Run(context.Background(), w)
			if err != nil {
				return fmt.Errorf("execute workflow: %w", err)
			}

			fmt.Printf("\nWorkflow completed: %s\n", result.Status)
			if result.Error != "" {
				fmt.Printf("Error: %s\n", result.Error)
			}
			for stepID, stepResult := range result.StepResults {
				fmt.Printf("\n--- Step: %s ---\n%s\n", stepID, truncate(stepResult, 500))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&schedule, "schedule", "", "cron schedule to register (overrides YAML)")
	return cmd
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
			defer boot.DBClient.Close()

			engine := initEngine(boot)
			if engine == nil {
				return fmt.Errorf("workflow engine is not enabled")
			}

			runs, err := engine.ListRuns(context.Background(), limit)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}

			if len(runs) == 0 {
				fmt.Println("No workflow runs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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
			defer boot.DBClient.Close()

			engine := initEngine(boot)
			if engine == nil {
				return fmt.Errorf("workflow engine is not enabled")
			}

			status, err := engine.Status(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}

			fmt.Printf("Run ID:    %s\n", status.RunID)
			fmt.Printf("Workflow:  %s\n", status.WorkflowName)
			fmt.Printf("Status:    %s\n", status.Status)
			fmt.Printf("Progress:  %d/%d steps\n", status.CompletedSteps, status.TotalSteps)

			if len(status.StepStatuses) > 0 {
				fmt.Println("\nSteps:")
				for _, s := range status.StepStatuses {
					errInfo := ""
					if s.Error != "" {
						errInfo = " (" + truncate(s.Error, 40) + ")"
					}
					fmt.Printf("  %-20s  %-12s  agent=%-15s%s\n",
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
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			engine := initEngine(boot)
			if engine == nil {
				return fmt.Errorf("workflow engine is not enabled")
			}

			if err := engine.Cancel(args[0]); err != nil {
				return fmt.Errorf("cancel workflow: %w", err)
			}

			fmt.Printf("Workflow run %s cancelled.\n", args[0])
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
			defer boot.DBClient.Close()

			engine := initEngine(boot)
			if engine == nil {
				return fmt.Errorf("workflow engine is not enabled")
			}

			runs, err := engine.ListRuns(context.Background(), limit)
			if err != nil {
				return fmt.Errorf("list runs: %w", err)
			}

			if len(runs) == 0 {
				fmt.Println("No workflow history found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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
	state := workflow.NewStateStore(boot.DBClient, lg.Sugar())
	return workflow.NewEngine(nil, state, nil,
		boot.Config.Workflow.MaxConcurrentSteps,
		boot.Config.Workflow.DefaultTimeout,
		lg.Sugar())
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.DateTime)
}
