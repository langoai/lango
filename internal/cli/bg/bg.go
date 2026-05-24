// Package bg provides CLI commands for background task management.
package bg

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/clihttp"
)

// Client is the narrow port used by bg commands.
type Client interface {
	List(context.Context) ([]Task, error)
	Status(context.Context, string) (Task, error)
	Result(context.Context, string) (string, error)
	Cancel(context.Context, string) error
}

// ClientProvider loads a background task client lazily for command execution.
type ClientProvider func() (Client, error)

// Task is the stable CLI/API representation of a background task.
type Task struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Prompt        string `json:"prompt,omitempty"`
	OriginChannel string `json:"originChannel,omitempty"`
	OriginSession string `json:"originSession,omitempty"`
	StartedAt     string `json:"startedAt,omitempty"`
	CompletedAt   string `json:"completedAt,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Error         string `json:"error,omitempty"`
	Result        string `json:"result,omitempty"`
}

type inProcessClient struct {
	managerProvider func() (*background.Manager, error)
}

// NewInProcessClientProvider adapts an in-process background manager provider.
func NewInProcessClientProvider(managerProvider func() (*background.Manager, error)) ClientProvider {
	return func() (Client, error) {
		return inProcessClient{managerProvider: managerProvider}, nil
	}
}

func (c inProcessClient) manager() (*background.Manager, error) {
	if c.managerProvider == nil {
		return nil, fmt.Errorf("background manager provider is not configured")
	}
	return c.managerProvider()
}

func (c inProcessClient) List(context.Context) ([]Task, error) {
	mgr, err := c.manager()
	if err != nil {
		return nil, err
	}
	snapshots := mgr.List()
	tasks := make([]Task, 0, len(snapshots))
	for _, snap := range snapshots {
		tasks = append(tasks, taskFromSnapshot(snap))
	}
	return tasks, nil
}

func (c inProcessClient) Status(_ context.Context, id string) (Task, error) {
	mgr, err := c.manager()
	if err != nil {
		return Task{}, err
	}
	snap, err := mgr.Status(id)
	if err != nil {
		return Task{}, err
	}
	return taskFromSnapshot(*snap), nil
}

func (c inProcessClient) Result(_ context.Context, id string) (string, error) {
	mgr, err := c.manager()
	if err != nil {
		return "", err
	}
	return mgr.Result(id)
}

func (c inProcessClient) Cancel(_ context.Context, id string) error {
	mgr, err := c.manager()
	if err != nil {
		return err
	}
	return mgr.Cancel(id)
}

type gatewayClient struct {
	addr string
}

// NewGatewayClient creates a gateway-backed background task client.
func NewGatewayClient(addr string) Client {
	return gatewayClient{addr: strings.TrimRight(addr, "/")}
}

type taskListResponse struct {
	Tasks []Task `json:"tasks"`
}

type taskStatusResponse struct {
	Task Task `json:"task"`
}

type taskResultResponse struct {
	Result string `json:"result"`
}

func (c gatewayClient) List(ctx context.Context) ([]Task, error) {
	var out taskListResponse
	if err := clihttp.FetchJSONContext(ctx, c.addr, "/api/bg/tasks", &out); err != nil {
		return nil, err
	}
	return out.Tasks, nil
}

func (c gatewayClient) Status(ctx context.Context, id string) (Task, error) {
	var out taskStatusResponse
	if err := clihttp.FetchJSONContext(ctx, c.addr, "/api/bg/tasks/"+url.PathEscape(id), &out); err != nil {
		return Task{}, err
	}
	return out.Task, nil
}

func (c gatewayClient) Result(ctx context.Context, id string) (string, error) {
	var out taskResultResponse
	if err := clihttp.FetchJSONContext(ctx, c.addr, "/api/bg/tasks/"+url.PathEscape(id)+"/result", &out); err != nil {
		return "", err
	}
	return out.Result, nil
}

func (c gatewayClient) Cancel(ctx context.Context, id string) error {
	var out struct {
		ID        string `json:"id"`
		Cancelled bool   `json:"cancelled"`
	}
	return clihttp.PostJSONContext(ctx, c.addr, "/api/bg/tasks/"+url.PathEscape(id)+"/cancel", map[string]string{}, &out)
}

// NewBgCmd creates the bg (background) command.
func NewBgCmd(clientProvider ClientProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bg",
		Short: "Manage background tasks",
		Long: `View, cancel, and retrieve results of background tasks.

These subcommands operate through a gateway-backed or in-process background task client.
Root CLI usage talks to a running Lango gateway; embedded callers can provide an
in-process manager adapter directly.`,
	}

	cmd.PersistentFlags().String("output", "table", "Output format: table or json")

	cmd.AddCommand(newBgListCmd(clientProvider))
	cmd.AddCommand(newBgStatusCmd(clientProvider))
	cmd.AddCommand(newBgCancelCmd(clientProvider))
	cmd.AddCommand(newBgResultCmd(clientProvider))

	return cmd
}

// NewGatewayCmd creates a root CLI bg command backed by the gateway API.
func NewGatewayCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var addr string
	cmd := NewBgCmd(func() (Client, error) {
		resolved, err := resolveGatewayAddr(addr, bootLoader)
		if err != nil {
			return nil, err
		}
		return NewGatewayClient(resolved), nil
	})
	cmd.PersistentFlags().StringVar(&addr, "addr", "", "Gateway address (default: config server host/port)")
	return cmd
}

func resolveGatewayAddr(addr string, bootLoader func() (*bootstrap.Result, error)) (string, error) {
	if trimmed := strings.TrimSpace(addr); trimmed != "" {
		return strings.TrimRight(trimmed, "/"), nil
	}
	if bootLoader == nil {
		return "", fmt.Errorf("gateway address is required when no config loader is available")
	}
	boot, err := bootLoader()
	if err != nil {
		return "", fmt.Errorf("load config for gateway address: %w", err)
	}
	if boot == nil || boot.Config == nil {
		return "", fmt.Errorf("load config for gateway address: config is unavailable")
	}
	defer func() { _ = boot.Close() }()
	return clihttp.ResolveGatewayAddr("", boot.Config), nil
}

func newBgListCmd(cp ClientProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List background tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cp()
			if err != nil {
				return fmt.Errorf("get background client: %w", err)
			}

			tasks, err := client.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}
			if printed, err := printJSONIfRequested(cmd, taskListResponse{Tasks: tasks}); err != nil {
				return err
			} else if printed {
				return nil
			}
			if len(tasks) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No background tasks.")
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tPROMPT\tSTARTED\tDURATION")
			for _, t := range tasks {
				prompt := t.Prompt
				if len(prompt) > 50 {
					prompt = prompt[:47] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					shortID(t.ID), t.Status, prompt,
					fallbackDash(t.StartedAt), fallbackDash(t.Duration))
			}
			return w.Flush()
		},
	}
}

func newBgStatusCmd(cp ClientProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Short: "Show background task status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cp()
			if err != nil {
				return fmt.Errorf("get background client: %w", err)
			}

			task, err := client.Status(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get status: %w", err)
			}
			if printed, err := printJSONIfRequested(cmd, taskStatusResponse{Task: task}); err != nil {
				return err
			} else if printed {
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", task.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:  %s\n", task.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Prompt:  %s\n", task.Prompt)
			fmt.Fprintf(cmd.OutOrStdout(), "Origin:  %s (session: %s)\n", task.OriginChannel, task.OriginSession)
			fmt.Fprintf(cmd.OutOrStdout(), "Started: %s\n", fallbackDash(task.StartedAt))
			if task.CompletedAt != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Completed: %s\n", task.CompletedAt)
				fmt.Fprintf(cmd.OutOrStdout(), "Duration: %s\n", fallbackDash(task.Duration))
			}
			if task.Error != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Error: %s\n", task.Error)
			}
			if task.Result != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "\nResult:\n%s\n", task.Result)
			}
			return nil
		},
	}
}

func newBgCancelCmd(cp ClientProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a running background task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cp()
			if err != nil {
				return fmt.Errorf("get background client: %w", err)
			}

			if err := client.Cancel(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("cancel task: %w", err)
			}
			if printed, err := printJSONIfRequested(cmd, map[string]interface{}{"id": args[0], "cancelled": true}); err != nil {
				return err
			} else if printed {
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Task %s cancelled.\n", args[0])
			return nil
		},
	}
}

func newBgResultCmd(cp ClientProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "result <id>",
		Short: "Show completed task result",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cp()
			if err != nil {
				return fmt.Errorf("get background client: %w", err)
			}

			result, err := client.Result(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get result: %w", err)
			}
			if printed, err := printJSONIfRequested(cmd, taskResultResponse{Result: result}); err != nil {
				return err
			} else if printed {
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), result)
			return nil
		},
	}
}

func printJSONIfRequested(cmd *cobra.Command, v interface{}) (bool, error) {
	format, err := clihttp.ResolveTableOrJSONOutput(cmd)
	if err != nil {
		return false, err
	}
	if format != "json" {
		return false, nil
	}
	return true, clihttp.PrintJSON(cmd.OutOrStdout(), v)
}

func taskFromSnapshot(s background.TaskSnapshot) Task {
	return Task{
		ID:            s.ID,
		Status:        s.Status.String(),
		Prompt:        s.Prompt,
		OriginChannel: s.OriginChannel,
		OriginSession: s.OriginSession,
		StartedAt:     formatTime(s.StartedAt),
		CompletedAt:   formatTime(s.CompletedAt),
		Duration:      taskDuration(s),
		Error:         s.Error,
		Result:        s.Result,
	}
}

func taskDuration(s background.TaskSnapshot) string {
	if !s.CompletedAt.IsZero() {
		return s.CompletedAt.Sub(s.StartedAt).Truncate(time.Millisecond).String()
	}
	if !s.StartedAt.IsZero() {
		return time.Since(s.StartedAt).Truncate(time.Second).String() + " (running)"
	}
	return ""
}

func fallbackDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}
