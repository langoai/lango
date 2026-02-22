// Package cron provides CLI commands for cron job management.
package cron

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/ent"
)

// NewCronCmd creates the cron command with lazy bootstrap loading.
func NewCronCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cron",
		Short: "Manage scheduled cron jobs",
		Long:  "Create, list, pause, resume, and delete scheduled tasks that run automatically.",
	}

	cmd.AddCommand(newAddCmd(bootLoader))
	cmd.AddCommand(newListCmd(bootLoader))
	cmd.AddCommand(newDeleteCmd(bootLoader))
	cmd.AddCommand(newPauseCmd(bootLoader))
	cmd.AddCommand(newResumeCmd(bootLoader))
	cmd.AddCommand(newHistoryCmd(bootLoader))

	return cmd
}

func initStore(boot *bootstrap.Result) cron.Store {
	return cron.NewEntStore(boot.DBClient)
}

func newAddCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		name      string
		schedule  string
		every     string
		at        string
		prompt    string
		deliverTo []string
		isolated  bool
		timezone  string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new cron job",
		Long: `Add a new scheduled cron job.

Examples:
  lango cron add --name "news" --schedule "0 9 * * *" --prompt "Summarize today's news" --deliver slack
  lango cron add --name "check" --every 1h --prompt "Check server status" --isolated
  lango cron add --name "meeting" --at "2026-02-20T15:00:00" --prompt "Prepare meeting notes"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if prompt == "" {
				return fmt.Errorf("--prompt is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Determine schedule type
			var scheduleType, scheduleVal string
			count := 0
			if schedule != "" {
				scheduleType = "cron"
				scheduleVal = schedule
				count++
			}
			if every != "" {
				scheduleType = "every"
				scheduleVal = every
				count++
			}
			if at != "" {
				scheduleType = "at"
				scheduleVal = at
				count++
			}
			if count == 0 {
				return fmt.Errorf("one of --schedule, --every, or --at is required")
			}
			if count > 1 {
				return fmt.Errorf("only one of --schedule, --every, or --at may be specified")
			}

			sessionMode := "main"
			if isolated {
				sessionMode = "isolated"
			}

			if timezone == "" {
				timezone = "UTC"
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)

			job := cron.Job{
				ID:           uuid.New().String(),
				Name:         name,
				ScheduleType: scheduleType,
				Schedule:     scheduleVal,
				Prompt:       prompt,
				SessionMode:  sessionMode,
				DeliverTo:    deliverTo,
				Timezone:     timezone,
				Enabled:      true,
				CreatedAt:    time.Now(),
			}

			if err := store.Create(context.Background(), job); err != nil {
				return fmt.Errorf("create job: %w", err)
			}

			fmt.Printf("Cron job %q created (id: %s)\n", name, job.ID)
			fmt.Printf("  Schedule: %s %s\n", scheduleType, scheduleVal)
			fmt.Printf("  Prompt: %s\n", truncate(prompt, 80))
			if len(deliverTo) > 0 {
				fmt.Printf("  Deliver to: %v\n", deliverTo)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "job name (required)")
	cmd.Flags().StringVar(&schedule, "schedule", "", "cron expression (e.g. '0 9 * * *')")
	cmd.Flags().StringVar(&every, "every", "", "interval (e.g. '1h', '30m')")
	cmd.Flags().StringVar(&at, "at", "", "one-time execution (ISO8601: '2026-02-20T15:00:00')")
	cmd.Flags().StringVar(&prompt, "prompt", "", "prompt to execute (required)")
	cmd.Flags().StringSliceVar(&deliverTo, "deliver", nil, "channels to deliver results (e.g. slack,telegram)")
	cmd.Flags().BoolVar(&isolated, "isolated", false, "run in isolated session")
	cmd.Flags().StringVar(&timezone, "timezone", "", "timezone (default: config or UTC)")

	return cmd
}

func newListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all cron jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)
			jobs, err := store.List(context.Background())
			if err != nil {
				return fmt.Errorf("list jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("No cron jobs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSCHEDULE\tENABLED\tLAST RUN\tNEXT RUN")
			for _, j := range jobs {
				lastRun := "-"
				if j.LastRunAt != nil {
					lastRun = j.LastRunAt.Format(time.DateTime)
				}
				nextRun := "-"
				if j.NextRunAt != nil {
					nextRun = j.NextRunAt.Format(time.DateTime)
				}
				enabled := "yes"
				if !j.Enabled {
					enabled = "no"
				}
				fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\t%s\n",
					shortID(j.ID), j.Name, j.ScheduleType, j.Schedule,
					enabled, lastRun, nextRun)
			}
			return w.Flush()
		},
	}
}

func newDeleteCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id-or-name>",
		Short: "Delete a cron job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)
			id, err := resolveJobID(context.Background(), store, args[0])
			if err != nil {
				return err
			}

			if err := store.Delete(context.Background(), id); err != nil {
				return fmt.Errorf("delete job: %w", err)
			}

			fmt.Printf("Cron job %q deleted.\n", args[0])
			return nil
		},
	}
}

func newPauseCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id-or-name>",
		Short: "Pause a cron job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)
			id, err := resolveJobID(context.Background(), store, args[0])
			if err != nil {
				return err
			}

			job, err := store.Get(context.Background(), id)
			if err != nil {
				return fmt.Errorf("get job: %w", err)
			}
			job.Enabled = false
			if err := store.Update(context.Background(), *job); err != nil {
				return fmt.Errorf("update job: %w", err)
			}

			fmt.Printf("Cron job %q paused.\n", args[0])
			return nil
		},
	}
}

func newResumeCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id-or-name>",
		Short: "Resume a paused cron job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)
			id, err := resolveJobID(context.Background(), store, args[0])
			if err != nil {
				return err
			}

			job, err := store.Get(context.Background(), id)
			if err != nil {
				return fmt.Errorf("get job: %w", err)
			}
			job.Enabled = true
			if err := store.Update(context.Background(), *job); err != nil {
				return fmt.Errorf("update job: %w", err)
			}

			fmt.Printf("Cron job %q resumed.\n", args[0])
			return nil
		},
	}
}

func newHistoryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history [id-or-name]",
		Short: "Show cron job execution history",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			store := initStore(boot)

			var entries []cron.HistoryEntry
			if len(args) > 0 {
				id, err := resolveJobID(context.Background(), store, args[0])
				if err != nil {
					return err
				}
				entries, err = store.ListHistory(context.Background(), id, limit)
				if err != nil {
					return fmt.Errorf("list history: %w", err)
				}
			} else {
				entries, err = store.ListAllHistory(context.Background(), limit)
				if err != nil {
					return fmt.Errorf("list history: %w", err)
				}
			}

			if len(entries) == 0 {
				fmt.Println("No execution history found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "JOB\tSTATUS\tSTARTED\tDURATION\tRESULT")
			for _, e := range entries {
				duration := "-"
				if e.CompletedAt != nil {
					duration = e.CompletedAt.Sub(e.StartedAt).Truncate(time.Millisecond).String()
				}
				result := truncate(e.Result, 60)
				if e.Status == "failed" && e.ErrorMessage != "" {
					result = "ERR: " + truncate(e.ErrorMessage, 55)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					e.JobName, e.Status, e.StartedAt.Format(time.DateTime),
					duration, result)
			}
			return w.Flush()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "maximum entries to show")
	return cmd
}

// resolveJobID tries to find a job by UUID or by name.
func resolveJobID(ctx context.Context, store cron.Store, idOrName string) (string, error) {
	// Try as UUID first
	if _, err := uuid.Parse(idOrName); err == nil {
		job, err := store.Get(ctx, idOrName)
		if err == nil && job != nil {
			return job.ID, nil
		}
	}

	// Try by name
	job, err := store.GetByName(ctx, idOrName)
	if err != nil {
		return "", fmt.Errorf("job %q not found: %w", idOrName, err)
	}
	return job.ID, nil
}

// resolveJobID for ent client (unused, keeping store-based approach)
var _ = (*ent.Client)(nil)

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
