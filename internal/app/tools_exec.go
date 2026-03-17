package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/supervisor"
	execpkg "github.com/langoai/lango/internal/tools/exec"
)

// BlockedResult is the structured response returned when a command is blocked
// by security guards (CLI guard, path guard, etc.).
type BlockedResult struct {
	Blocked bool   `json:"blocked"`
	Message string `json:"message"`
}

func buildExecTools(sv *supervisor.Supervisor, automationAvailable map[string]bool, guard *execpkg.CommandGuard) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "exec",
			Description: "Execute shell commands",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("command", "The shell command to execute").
				Required("command").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, ok := params["command"].(string)
				if !ok {
					return nil, fmt.Errorf("missing command parameter")
				}
				if msg := blockLangoExec(cmd, automationAvailable); msg != "" {
					return &BlockedResult{Blocked: true, Message: msg}, nil
				}
				if msg := blockProtectedPaths(cmd, guard); msg != "" {
					return &BlockedResult{Blocked: true, Message: msg}, nil
				}
				return sv.ExecuteTool(ctx, cmd)
			},
		},
		{
			Name:        "exec_bg",
			Description: "Execute a shell command in the background",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("command", "The shell command to execute").
				Required("command").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, ok := params["command"].(string)
				if !ok {
					return nil, fmt.Errorf("missing command parameter")
				}
				if msg := blockLangoExec(cmd, automationAvailable); msg != "" {
					return &BlockedResult{Blocked: true, Message: msg}, nil
				}
				if msg := blockProtectedPaths(cmd, guard); msg != "" {
					return &BlockedResult{Blocked: true, Message: msg}, nil
				}
				return sv.StartBackground(cmd)
			},
		},
		{
			Name:        "exec_status",
			Description: "Check the status of a background process",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: agent.Schema().
				Str("id", "The background process ID returned by exec_bg").
				Required("id").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, ok := params["id"].(string)
				if !ok {
					return nil, fmt.Errorf("missing id parameter")
				}
				return sv.GetBackgroundStatus(id)
			},
		},
		{
			Name:        "exec_stop",
			Description: "Stop a background process",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("id", "The background process ID returned by exec_bg").
				Required("id").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, ok := params["id"].(string)
				if !ok {
					return nil, fmt.Errorf("missing id parameter")
				}
				return nil, sv.StopBackground(id)
			},
		},
	}
}
