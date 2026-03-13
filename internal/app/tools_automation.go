package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/background"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/workflow"
)

// buildCronTools creates tools for managing scheduled cron jobs.
func buildCronTools(scheduler *cronpkg.Scheduler, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "cron_add",
			Description: "Create a new scheduled cron job that runs an agent prompt on a recurring schedule",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":          map[string]interface{}{"type": "string", "description": "Unique name for the cron job"},
					"schedule_type": map[string]interface{}{"type": "string", "description": "Schedule type: cron (crontab), every (interval), or at (one-time)", "enum": []string{"cron", "every", "at"}},
					"schedule":      map[string]interface{}{"type": "string", "description": "Schedule value: crontab expr for cron, Go duration for every (e.g. 1h30m), RFC3339 datetime for at"},
					"prompt":        map[string]interface{}{"type": "string", "description": "The prompt to execute on each run"},
					"session_mode":  map[string]interface{}{"type": "string", "description": "Session mode: isolated (new session each run) or main (shared session)", "enum": []string{"isolated", "main"}},
					"deliver_to":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Channels to deliver results to (e.g. telegram:CHAT_ID, discord:CHANNEL_ID, slack:CHANNEL_ID)"},
					"timeout":       map[string]interface{}{"type": "string", "description": "Per-job timeout as Go duration (e.g. 10m, 1h). Overrides the default job timeout."},
				},
				"required": []string{"name", "schedule_type", "schedule", "prompt"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				name, err := toolparam.RequireString(params, "name")
				if err != nil {
					return nil, err
				}
				scheduleType, err := toolparam.RequireString(params, "schedule_type")
				if err != nil {
					return nil, err
				}
				schedule, err := toolparam.RequireString(params, "schedule")
				if err != nil {
					return nil, err
				}
				prompt, err := toolparam.RequireString(params, "prompt")
				if err != nil {
					return nil, err
				}
				sessionMode := toolparam.OptionalString(params, "session_mode", "isolated")

				deliverTo := toolparam.StringSlice(params, "deliver_to")

				// Auto-detect channel from session context.
				if len(deliverTo) == 0 {
					if ch := detectChannelFromContext(ctx); ch != "" {
						deliverTo = []string{ch}
					}
				}
				// Fall back to config default.
				if len(deliverTo) == 0 && len(defaultDeliverTo) > 0 {
					deliverTo = make([]string, len(defaultDeliverTo))
					copy(deliverTo, defaultDeliverTo)
				}

				var timeout time.Duration
				if t, ok := params["timeout"].(string); ok && t != "" {
					var parseErr error
					timeout, parseErr = time.ParseDuration(t)
					if parseErr != nil {
						return nil, fmt.Errorf("parse timeout %q: %w", t, parseErr)
					}
				}

				job := cronpkg.Job{
					Name:         name,
					ScheduleType: scheduleType,
					Schedule:     schedule,
					Prompt:       prompt,
					SessionMode:  sessionMode,
					DeliverTo:    deliverTo,
					Timeout:      timeout,
					Enabled:      true,
				}

				updated, err := scheduler.AddJob(ctx, job)
				if err != nil {
					return nil, fmt.Errorf("add cron job: %w", err)
				}

				status := "created"
				verb := "created"
				if updated {
					status = "updated"
					verb = "updated"
				}

				return map[string]interface{}{
					"status":  status,
					"name":    name,
					"message": fmt.Sprintf("Cron job '%s' %s with schedule %s=%s", name, verb, scheduleType, schedule),
				}, nil
			},
		},
		{
			Name:        "cron_list",
			Description: "List all registered cron jobs with their schedules and status",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				jobs, err := scheduler.ListJobs(ctx)
				if err != nil {
					return nil, fmt.Errorf("list cron jobs: %w", err)
				}
				return map[string]interface{}{"jobs": jobs, "count": len(jobs)}, nil
			},
		},
		{
			Name:        "cron_pause",
			Description: "Pause a cron job so it no longer fires on schedule",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID to pause"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				if err := scheduler.PauseJob(ctx, id); err != nil {
					return nil, fmt.Errorf("pause cron job: %w", err)
				}
				return map[string]interface{}{"status": "paused", "id": id}, nil
			},
		},
		{
			Name:        "cron_resume",
			Description: "Resume a paused cron job",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID to resume"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				if err := scheduler.ResumeJob(ctx, id); err != nil {
					return nil, fmt.Errorf("resume cron job: %w", err)
				}
				return map[string]interface{}{"status": "resumed", "id": id}, nil
			},
		},
		{
			Name:        "cron_remove",
			Description: "Permanently remove a cron job",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID to remove"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				if err := scheduler.RemoveJob(ctx, id); err != nil {
					return nil, fmt.Errorf("remove cron job: %w", err)
				}
				return map[string]interface{}{"status": "removed", "id": id}, nil
			},
		},
		{
			Name:        "cron_history",
			Description: "View execution history for cron jobs",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"job_id": map[string]interface{}{"type": "string", "description": "Filter by job ID (omit for all jobs)"},
					"limit":  map[string]interface{}{"type": "integer", "description": "Maximum entries to return (default: 20)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				jobID := toolparam.OptionalString(params, "job_id", "")
				limit := toolparam.OptionalInt(params, "limit", 20)

				var entries []cronpkg.HistoryEntry
				var err error
				if jobID != "" {
					entries, err = scheduler.History(ctx, jobID, limit)
				} else {
					entries, err = scheduler.AllHistory(ctx, limit)
				}
				if err != nil {
					return nil, fmt.Errorf("cron history: %w", err)
				}
				return map[string]interface{}{"entries": entries, "count": len(entries)}, nil
			},
		},
	}
}

// buildBackgroundTools creates tools for managing background tasks.
func buildBackgroundTools(mgr *background.Manager, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "bg_submit",
			Description: "Submit a prompt for asynchronous background execution",
			SafetyLevel: agent.SafetyLevelModerate,
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
					channel = detectChannelFromContext(ctx)
				}
				// Fall back to config default.
				if channel == "" && len(defaultDeliverTo) > 0 {
					channel = defaultDeliverTo[0]
				}

				sessionKey := session.SessionKeyFromContext(ctx)

				taskID, err := mgr.Submit(ctx, prompt, background.Origin{
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

// buildWorkflowTools creates tools for executing and managing workflows.
func buildWorkflowTools(engine *workflow.Engine, stateDir string, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "workflow_run",
			Description: "Execute a workflow from a YAML file path or inline YAML content",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path":    map[string]interface{}{"type": "string", "description": "Path to a .flow.yaml workflow file"},
					"yaml_content": map[string]interface{}{"type": "string", "description": "Inline YAML workflow definition (alternative to file_path)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				filePath := toolparam.OptionalString(params, "file_path", "")
				yamlContent := toolparam.OptionalString(params, "yaml_content", "")

				if filePath == "" && yamlContent == "" {
					return nil, fmt.Errorf("either file_path or yaml_content is required")
				}

				var w *workflow.Workflow
				var err error
				if filePath != "" {
					w, err = workflow.ParseFile(filePath)
				} else {
					w, err = workflow.Parse([]byte(yamlContent))
				}
				if err != nil {
					return nil, fmt.Errorf("parse workflow: %w", err)
				}

				// Auto-detect delivery channel from session context.
				if len(w.DeliverTo) == 0 {
					if ch := detectChannelFromContext(ctx); ch != "" {
						w.DeliverTo = []string{ch}
					}
				}
				// Fall back to config default.
				if len(w.DeliverTo) == 0 && len(defaultDeliverTo) > 0 {
					w.DeliverTo = make([]string, len(defaultDeliverTo))
					copy(w.DeliverTo, defaultDeliverTo)
				}

				runID, err := engine.RunAsync(ctx, w)
				if err != nil {
					return nil, fmt.Errorf("run workflow: %w", err)
				}

				return map[string]interface{}{
					"run_id":  runID,
					"status":  "running",
					"message": fmt.Sprintf("Workflow '%s' started. Use workflow_status to check progress.", w.Name),
				}, nil
			},
		},
		{
			Name:        "workflow_status",
			Description: "Check the current status and progress of a workflow execution",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"run_id": map[string]interface{}{"type": "string", "description": "The workflow run ID"},
				},
				"required": []string{"run_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				runID, err := toolparam.RequireString(params, "run_id")
				if err != nil {
					return nil, err
				}
				status, err := engine.Status(ctx, runID)
				if err != nil {
					return nil, fmt.Errorf("workflow status: %w", err)
				}
				return status, nil
			},
		},
		{
			Name:        "workflow_list",
			Description: "List recent workflow executions",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer", "description": "Maximum runs to return (default: 20)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				limit := toolparam.OptionalInt(params, "limit", 20)
				runs, err := engine.ListRuns(ctx, limit)
				if err != nil {
					return nil, fmt.Errorf("list workflow runs: %w", err)
				}
				return map[string]interface{}{"runs": runs, "count": len(runs)}, nil
			},
		},
		{
			Name:        "workflow_cancel",
			Description: "Cancel a running workflow execution",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"run_id": map[string]interface{}{"type": "string", "description": "The workflow run ID to cancel"},
				},
				"required": []string{"run_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				runID, err := toolparam.RequireString(params, "run_id")
				if err != nil {
					return nil, err
				}
				if err := engine.Cancel(runID); err != nil {
					return nil, fmt.Errorf("cancel workflow: %w", err)
				}
				return map[string]interface{}{"status": "cancelled", "run_id": runID}, nil
			},
		},
		{
			Name:        "workflow_save",
			Description: "Save a workflow YAML definition to the workflows directory for future use",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":         map[string]interface{}{"type": "string", "description": "Workflow name (used as filename: name.flow.yaml)"},
					"yaml_content": map[string]interface{}{"type": "string", "description": "The YAML workflow definition"},
				},
				"required": []string{"name", "yaml_content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				name, err := toolparam.RequireString(params, "name")
				if err != nil {
					return nil, err
				}
				yamlContent, err := toolparam.RequireString(params, "yaml_content")
				if err != nil {
					return nil, err
				}

				// Validate the YAML before saving.
				w, err := workflow.Parse([]byte(yamlContent))
				if err != nil {
					return nil, fmt.Errorf("parse workflow YAML: %w", err)
				}
				if err := workflow.Validate(w); err != nil {
					return nil, fmt.Errorf("validate workflow: %w", err)
				}

				dir := stateDir
				if dir == "" {
					if home, err := os.UserHomeDir(); err == nil {
						dir = filepath.Join(home, ".lango", "workflows")
					} else {
						return nil, fmt.Errorf("determine workflows directory: %w", err)
					}
				}

				if err := os.MkdirAll(dir, 0o755); err != nil {
					return nil, fmt.Errorf("create workflows directory: %w", err)
				}

				filePath := filepath.Join(dir, name+".flow.yaml")
				if err := os.WriteFile(filePath, []byte(yamlContent), 0o644); err != nil {
					return nil, fmt.Errorf("write workflow file: %w", err)
				}

				return map[string]interface{}{
					"status":    "saved",
					"name":      name,
					"file_path": filePath,
					"message":   fmt.Sprintf("Workflow '%s' saved to %s", name, filePath),
				}, nil
			},
		},
	}
}
