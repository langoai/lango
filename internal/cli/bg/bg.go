// Package bg provides CLI commands for background task management.
package bg

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/background"
)

// NewBgCmd creates the bg (background) command.
// The manager is provided lazily since it only exists when the server is running.
func NewBgCmd(managerProvider func() (*background.Manager, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bg",
		Short: "Manage background tasks",
		Long:  "View, cancel, and retrieve results of background tasks.",
	}

	cmd.AddCommand(newBgListCmd(managerProvider))
	cmd.AddCommand(newBgStatusCmd(managerProvider))
	cmd.AddCommand(newBgCancelCmd(managerProvider))
	cmd.AddCommand(newBgResultCmd(managerProvider))

	return cmd
}

func newBgListCmd(mp func() (*background.Manager, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List background tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := mp()
			if err != nil {
				return fmt.Errorf("get manager: %w", err)
			}

			tasks := mgr.List()
			if len(tasks) == 0 {
				fmt.Println("No background tasks.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tPROMPT\tSTARTED\tDURATION")
			for _, t := range tasks {
				duration := "-"
				if !t.CompletedAt.IsZero() {
					duration = t.CompletedAt.Sub(t.StartedAt).Truncate(time.Millisecond).String()
				} else if !t.StartedAt.IsZero() {
					duration = time.Since(t.StartedAt).Truncate(time.Second).String() + " (running)"
				}
				prompt := t.Prompt
				if len(prompt) > 50 {
					prompt = prompt[:47] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					shortID(t.ID), t.Status.String(), prompt,
					formatTime(t.StartedAt), duration)
			}
			return w.Flush()
		},
	}
}

func newBgStatusCmd(mp func() (*background.Manager, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Short: "Show background task status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := mp()
			if err != nil {
				return fmt.Errorf("get manager: %w", err)
			}

			task, err := mgr.Status(args[0])
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}

			fmt.Printf("ID:      %s\n", task.ID)
			fmt.Printf("Status:  %s\n", task.Status.String())
			fmt.Printf("Prompt:  %s\n", task.Prompt)
			fmt.Printf("Origin:  %s (session: %s)\n", task.OriginChannel, task.OriginSession)
			fmt.Printf("Started: %s\n", formatTime(task.StartedAt))
			if !task.CompletedAt.IsZero() {
				fmt.Printf("Completed: %s\n", formatTime(task.CompletedAt))
				fmt.Printf("Duration: %s\n", task.CompletedAt.Sub(task.StartedAt).Truncate(time.Millisecond))
			}
			if task.Error != "" {
				fmt.Printf("Error: %s\n", task.Error)
			}
			if task.Result != "" {
				fmt.Printf("\nResult:\n%s\n", task.Result)
			}
			return nil
		},
	}
}

func newBgCancelCmd(mp func() (*background.Manager, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a running background task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := mp()
			if err != nil {
				return fmt.Errorf("get manager: %w", err)
			}

			if err := mgr.Cancel(args[0]); err != nil {
				return fmt.Errorf("cancel task: %w", err)
			}

			fmt.Printf("Task %s cancelled.\n", args[0])
			return nil
		},
	}
}

func newBgResultCmd(mp func() (*background.Manager, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "result <id>",
		Short: "Show completed task result",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := mp()
			if err != nil {
				return fmt.Errorf("get manager: %w", err)
			}

			result, err := mgr.Result(args[0])
			if err != nil {
				return fmt.Errorf("get result: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.DateTime)
}
