package exec

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
)

// BlockedResult is the structured response returned when a command is blocked
// by security guards (CLI guard, path guard, etc.).
type BlockedResult struct {
	Blocked bool   `json:"blocked"`
	Message string `json:"message"`
}

// GuardFunc is a function that checks a command and returns a non-empty
// guidance message if the command should be blocked.
type GuardFunc func(cmd string) string

// Executor abstracts the supervisor methods used by exec tools.
// *supervisor.Supervisor satisfies this interface.
type Executor interface {
	ExecuteTool(ctx context.Context, cmd string) (string, error)
	StartBackground(cmd string) (string, error)
	GetBackgroundStatus(id string) (map[string]interface{}, error)
	StopBackground(id string) error
}

// BuildTools creates exec agent tools backed by the given Executor.
// guardFns are command guard functions invoked before execution; if any
// returns a non-empty string the command is blocked.
func BuildTools(ex Executor, guardFns ...GuardFunc) []*agent.Tool {
	checkGuards := func(cmd string) *BlockedResult {
		for _, fn := range guardFns {
			if msg := fn(cmd); msg != "" {
				return &BlockedResult{Blocked: true, Message: msg}
			}
		}
		return nil
	}

	return []*agent.Tool{
		{
			Name:        "exec",
			Description: "Execute shell commands",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "execution",
				Activity:    agent.ActivityExecute,
				Aliases:     []string{"run", "shell", "bash", "command"},
				SearchHints: []string{"terminal", "command line", "sh"},
			},
			Parameters: agent.Schema().
				Str("command", "The shell command to execute").
				Required("command").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, err := toolparam.RequireString(params, "command")
				if err != nil {
					return nil, err
				}
				if br := checkGuards(cmd); br != nil {
					return br, nil
				}
				return ex.ExecuteTool(ctx, cmd)
			},
		},
		{
			Name:        "exec_bg",
			Description: "Execute a shell command in the background",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "execution",
				Activity:    agent.ActivityExecute,
				Aliases:     []string{"background_exec", "run_bg"},
				SearchHints: []string{"background", "async"},
			},
			Parameters: agent.Schema().
				Str("command", "The shell command to execute").
				Required("command").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, err := toolparam.RequireString(params, "command")
				if err != nil {
					return nil, err
				}
				if br := checkGuards(cmd); br != nil {
					return br, nil
				}
				return ex.StartBackground(cmd)
			},
		},
		{
			Name:        "exec_status",
			Description: "Check the status of a background process",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "execution",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
				Aliases:         []string{"job_status"},
				SearchHints:     []string{"background", "status"},
			},
			Parameters: agent.Schema().
				Str("id", "The background process ID returned by exec_bg").
				Required("id").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				return ex.GetBackgroundStatus(id)
			},
		},
		{
			Name:        "exec_stop",
			Description: "Stop a background process",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category:    "execution",
				Activity:    agent.ActivityManage,
				Aliases:     []string{"stop_job", "kill"},
				SearchHints: []string{"cancel", "stop"},
			},
			Parameters: agent.Schema().
				Str("id", "The background process ID returned by exec_bg").
				Required("id").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				return nil, ex.StopBackground(id)
			},
		},
	}
}
