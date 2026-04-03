package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/automation"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools for managing scheduled cron jobs.
func BuildTools(scheduler *Scheduler, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "cron_add",
			Description: "Create a new scheduled cron job that runs an agent prompt on a recurring schedule",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityManage,
			},
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
					if ch := automation.DetectChannelFromContext(ctx); ch != "" {
						deliverTo = []string{ch}
					}
				}
				// Fall back to config default.
				if len(deliverTo) == 0 && len(defaultDeliverTo) > 0 {
					deliverTo = make([]string, len(defaultDeliverTo))
					copy(deliverTo, defaultDeliverTo)
				}

				var timeout time.Duration
				if t := toolparam.OptionalString(params, "timeout", ""); t != "" {
					var parseErr error
					timeout, parseErr = time.ParseDuration(t)
					if parseErr != nil {
						return nil, fmt.Errorf("parse timeout %q: %w", t, parseErr)
					}
				}

				job := Job{
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
			Capability: agent.ToolCapability{
				Category:        "automation",
				Activity:        agent.ActivityManage,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID or name"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				nameOrID, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				id, err := scheduler.ResolveJobID(ctx, nameOrID)
				if err != nil {
					return nil, fmt.Errorf("pause cron job: %w", err)
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID or name"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				nameOrID, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				id, err := scheduler.ResolveJobID(ctx, nameOrID)
				if err != nil {
					return nil, fmt.Errorf("resume cron job: %w", err)
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "The cron job ID or name"},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				nameOrID, err := toolparam.RequireString(params, "id")
				if err != nil {
					return nil, err
				}
				id, err := scheduler.ResolveJobID(ctx, nameOrID)
				if err != nil {
					return nil, fmt.Errorf("remove cron job: %w", err)
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
			Capability: agent.ToolCapability{
				Category:        "automation",
				Activity:        agent.ActivityManage,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
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

				var entries []HistoryEntry
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
