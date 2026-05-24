// Package cron provides CLI commands for cron job management.
package cron

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/toolchain"
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
	if boot != nil && boot.Storage != nil {
		return boot.Storage.Cron()
	}
	return nil
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
		timeout   string
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

			if timezone == "" {
				timezone = "UTC"
			}
			var jobTimeout time.Duration
			if timeout != "" {
				var err error
				jobTimeout, err = time.ParseDuration(timeout)
				if err != nil {
					return fmt.Errorf("parse --timeout %q: %w", timeout, err)
				}
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := initStore(boot)
			sessionMode, err := cronSessionMode(cmd, boot.Config, isolated)
			if err != nil {
				return err
			}

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
				Timeout:      jobTimeout,
				CreatedAt:    time.Now(),
			}

			stored, updated, err := store.Upsert(context.Background(), job)
			if err != nil {
				return fmt.Errorf("upsert job: %w", err)
			}

			out := cmd.OutOrStdout()
			action := "created"
			if updated {
				action = "updated"
			}
			fmt.Fprintf(out, "Cron job %q %s (id: %s)\n", name, action, stored.ID)
			fmt.Fprintf(out, "  Schedule: %s %s\n", scheduleType, scheduleVal)
			fmt.Fprintf(out, "  Prompt: %s\n", truncate(prompt, 80))
			if len(deliverTo) > 0 {
				fmt.Fprintf(out, "  Deliver to: %v\n", deliverTo)
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
	cmd.Flags().StringSliceVar(&deliverTo, "deliver-to", nil, "channels to deliver results (alias for --deliver)")
	cmd.Flags().BoolVar(
		&isolated,
		"isolated",
		false,
		"override cron.defaultSessionMode for this job (true=isolated, false=main)",
	)
	cmd.Flags().StringVar(&timezone, "timezone", "", "timezone (default: config or UTC)")
	cmd.Flags().StringVar(&timeout, "timeout", "", "per-job timeout (Go duration, e.g. 5m, 1h30m)")

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
			defer boot.Close()

			store := initStore(boot)
			jobs, err := store.List(context.Background())
			if err != nil {
				return fmt.Errorf("list jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No cron jobs found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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
	var id string

	cmd := &cobra.Command{
		Use:   "delete <id-or-name>",
		Short: "Delete a cron job",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := requiredCronSelector(args, id)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := initStore(boot)
			resolvedID, err := resolveJobID(context.Background(), store, selector)
			if err != nil {
				return err
			}

			if err := store.Delete(context.Background(), resolvedID); err != nil {
				return fmt.Errorf("delete job: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cron job %q deleted.\n", selector)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "job ID or name")
	return cmd
}

func newPauseCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "pause <id-or-name>",
		Short: "Pause a cron job",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := requiredCronSelector(args, id)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := initStore(boot)
			resolvedID, err := resolveJobID(context.Background(), store, selector)
			if err != nil {
				return err
			}

			job, err := store.Get(context.Background(), resolvedID)
			if err != nil {
				return fmt.Errorf("get job: %w", err)
			}
			job.Enabled = false
			if err := store.Update(context.Background(), *job); err != nil {
				return fmt.Errorf("update job: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cron job %q paused.\n", selector)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "job ID or name")
	return cmd
}

func newResumeCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "resume <id-or-name>",
		Short: "Resume a paused cron job",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := requiredCronSelector(args, id)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := initStore(boot)
			resolvedID, err := resolveJobID(context.Background(), store, selector)
			if err != nil {
				return err
			}

			job, err := store.Get(context.Background(), resolvedID)
			if err != nil {
				return fmt.Errorf("get job: %w", err)
			}
			job.Enabled = true
			if err := store.Update(context.Background(), *job); err != nil {
				return fmt.Errorf("update job: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cron job %q resumed.\n", selector)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "job ID or name")
	return cmd
}

func newHistoryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		id    string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "history [id-or-name]",
		Short: "Show cron job execution history",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, hasSelector, err := optionalCronSelector(args, id)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			store := initStore(boot)

			var entries []cron.HistoryEntry
			if hasSelector {
				resolvedID, err := resolveJobID(context.Background(), store, selector)
				if err != nil {
					return err
				}
				entries, err = store.ListHistory(context.Background(), resolvedID, limit)
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
				fmt.Fprintln(cmd.OutOrStdout(), "No execution history found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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

	cmd.Flags().StringVar(&id, "id", "", "job ID or name")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "maximum entries to show")
	return cmd
}

func requiredCronSelector(args []string, id string) (string, error) {
	selector, ok, err := optionalCronSelector(args, id)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("requires <id-or-name> or --id")
	}
	return selector, nil
}

func optionalCronSelector(args []string, id string) (string, bool, error) {
	if len(args) > 0 && id != "" {
		return "", false, fmt.Errorf("provide either <id-or-name> or --id, not both")
	}
	if id != "" {
		return id, true, nil
	}
	if len(args) > 0 {
		return args[0], true, nil
	}
	return "", false, nil
}

func cronSessionMode(cmd *cobra.Command, cfg *config.Config, isolated bool) (string, error) {
	if cmd.Flags().Changed("isolated") {
		if isolated {
			return "isolated", nil
		}
		return "main", nil
	}

	sessionMode := "isolated"
	if cfg != nil && cfg.Cron.DefaultSessionMode != "" {
		sessionMode = cfg.Cron.DefaultSessionMode
	}
	if sessionMode != "isolated" && sessionMode != "main" {
		return "", fmt.Errorf("invalid cron.defaultSessionMode %q: expected isolated or main", sessionMode)
	}
	return sessionMode, nil
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
	return toolchain.Truncate(s, max)
}
