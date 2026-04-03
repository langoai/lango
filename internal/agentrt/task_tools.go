package agentrt

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTaskTools creates task management tools backed by the given TaskStore.
func BuildTaskTools(store TaskStore) []*agent.Tool {
	return []*agent.Tool{
		buildTaskCreate(store),
		buildTaskGet(store),
		buildTaskList(store),
		buildTaskUpdate(store),
	}
}

func buildTaskCreate(store TaskStore) *agent.Tool {
	return &agent.Tool{
		Name:        "task_create",
		Description: "Create a new task for structured task tracking",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("title", "The task title (required)").
			Str("description", "Optional task description").
			Str("parent_id", "Optional parent task ID for hierarchical tasks").
			Required("title").
			Build(),
		Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			title, err := toolparam.RequireString(params, "title")
			if err != nil {
				return nil, err
			}

			taskID, err := generateTaskID()
			if err != nil {
				return nil, err
			}

			now := time.Now()
			entry := &TaskEntry{
				ID:          taskID,
				Title:       title,
				Status:      "todo",
				Description: toolparam.OptionalString(params, "description", ""),
				ParentID:    toolparam.OptionalString(params, "parent_id", ""),
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := store.Create(entry); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"task_id": taskID,
				"status":  "todo",
			}, nil
		},
	}
}

func buildTaskGet(store TaskStore) *agent.Tool {
	return &agent.Tool{
		Name:        "task_get",
		Description: "Get a task by ID with full details",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("task_id", "The task ID to retrieve (required)").
			Required("task_id").
			Build(),
		Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			taskID, err := toolparam.RequireString(params, "task_id")
			if err != nil {
				return nil, err
			}

			entry, err := store.Get(taskID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"id":          entry.ID,
				"title":       entry.Title,
				"status":      entry.Status,
				"agent_id":    entry.AgentID,
				"parent_id":   entry.ParentID,
				"description": entry.Description,
				"created_at":  entry.CreatedAt.Format(time.RFC3339),
				"updated_at":  entry.UpdatedAt.Format(time.RFC3339),
			}, nil
		},
	}
}

func buildTaskList(store TaskStore) *agent.Tool {
	return &agent.Tool{
		Name:        "task_list",
		Description: "List tasks with optional status and parent filters",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("status", "Optional status filter: todo, in_progress, done, blocked").
			Str("parent_id", "Optional parent task ID filter").
			Build(),
		Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			statusFilter := toolparam.OptionalString(params, "status", "")
			parentFilter := toolparam.OptionalString(params, "parent_id", "")

			entries := store.List(statusFilter, parentFilter)

			tasks := make([]map[string]interface{}, 0, len(entries))
			for _, e := range entries {
				tasks = append(tasks, map[string]interface{}{
					"id":          e.ID,
					"title":       e.Title,
					"status":      e.Status,
					"agent_id":    e.AgentID,
					"parent_id":   e.ParentID,
					"description": e.Description,
					"created_at":  e.CreatedAt.Format(time.RFC3339),
					"updated_at":  e.UpdatedAt.Format(time.RFC3339),
				})
			}

			return map[string]interface{}{
				"tasks": tasks,
				"count": len(tasks),
			}, nil
		},
	}
}

func buildTaskUpdate(store TaskStore) *agent.Tool {
	return &agent.Tool{
		Name:        "task_update",
		Description: "Update a task's status and/or description",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("task_id", "The task ID to update (required)").
			Str("status", "New status: todo, in_progress, done, blocked").
			Str("description", "New description").
			Required("task_id").
			Build(),
		Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			taskID, err := toolparam.RequireString(params, "task_id")
			if err != nil {
				return nil, err
			}

			status := toolparam.OptionalString(params, "status", "")
			description := toolparam.OptionalString(params, "description", "")

			if err := store.Update(taskID, status, description); err != nil {
				return nil, err
			}

			entry, err := store.Get(taskID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"id":          entry.ID,
				"title":       entry.Title,
				"status":      entry.Status,
				"agent_id":    entry.AgentID,
				"parent_id":   entry.ParentID,
				"description": entry.Description,
				"created_at":  entry.CreatedAt.Format(time.RFC3339),
				"updated_at":  entry.UpdatedAt.Format(time.RFC3339),
			}, nil
		},
	}
}
