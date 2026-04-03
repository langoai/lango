package background

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/automation"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools for managing background tasks.
func BuildTools(mgr *Manager, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "bg_submit",
			Description: "Submit a prompt for asynchronous background execution",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt":  map[string]interface{}{"type": "string", "description": "The prompt to execute in the background"},
					"channel": map[string]interface{}{"type": "string", "description": "Channel to deliver results to (e.g. telegram:CHAT_ID, discord:CHANNEL_ID, slack:CHANNEL_ID)"},
				},
				"required": []string{"prompt"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				prompt, err := toolparam.RequireString(params, "prompt")
				if err != nil {
					return nil, err
				}
				channel := toolparam.OptionalString(params, "channel", "")

				// Auto-detect channel from session context.
				if channel == "" {
					channel = automation.DetectChannelFromContext(ctx)
				}
				// Fall back to config default.
				if channel == "" && len(defaultDeliverTo) > 0 {
					channel = defaultDeliverTo[0]
				}

				sessionKey := session.SessionKeyFromContext(ctx)

				taskID, err := mgr.Submit(ctx, prompt, Origin{
					Channel: channel,
					Session: sessionKey,
				})
				if err != nil {
					return nil, fmt.Errorf("submit background task: %w", err)
				}
				return map[string]interface{}{
					"status":  "submitted",
					"task_id": taskID,
					"message": "Task submitted for background execution",
				}, nil
			},
		},
		{
			Name:        "bg_status",
			Description: "Check the status of a background task",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "automation",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "string", "description": "The background task ID"},
				},
				"required": []string{"task_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "task_id")
				if err != nil {
					return nil, err
				}
				snap, err := mgr.Status(taskID)
				if err != nil {
					return nil, fmt.Errorf("background task status: %w", err)
				}
				return snap, nil
			},
		},
		{
			Name:        "bg_list",
			Description: "List all background tasks and their current status",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "automation",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				snapshots := mgr.List()
				return map[string]interface{}{"tasks": snapshots, "count": len(snapshots)}, nil
			},
		},
		{
			Name:        "bg_result",
			Description: "Retrieve the result of a completed background task",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "automation",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "string", "description": "The background task ID"},
				},
				"required": []string{"task_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "task_id")
				if err != nil {
					return nil, err
				}
				result, err := mgr.Result(taskID)
				if err != nil {
					return nil, fmt.Errorf("background task result: %w", err)
				}
				return map[string]interface{}{"task_id": taskID, "result": result}, nil
			},
		},
		{
			Name:        "bg_cancel",
			Description: "Cancel a pending or running background task",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "string", "description": "The background task ID to cancel"},
				},
				"required": []string{"task_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				taskID, err := toolparam.RequireString(params, "task_id")
				if err != nil {
					return nil, err
				}
				if err := mgr.Cancel(taskID); err != nil {
					return nil, fmt.Errorf("cancel background task: %w", err)
				}
				return map[string]interface{}{"status": "cancelled", "task_id": taskID}, nil
			},
		},
	}
}
