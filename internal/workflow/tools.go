package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/automation"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools for executing and managing workflows.
func BuildTools(engine *Engine, stateDir string, defaultDeliverTo []string) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "workflow_run",
			Description: "Execute a workflow from a YAML file path or inline YAML content",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
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

				var w *Workflow
				var err error
				if filePath != "" {
					w, err = ParseFile(filePath)
				} else {
					w, err = Parse([]byte(yamlContent))
				}
				if err != nil {
					return nil, fmt.Errorf("parse workflow: %w", err)
				}

				// Auto-detect delivery channel from session context.
				if len(w.DeliverTo) == 0 {
					if ch := automation.DetectChannelFromContext(ctx); ch != "" {
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
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
			Capability: agent.ToolCapability{
				Category: "automation",
				Activity: agent.ActivityExecute,
			},
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
				w, err := Parse([]byte(yamlContent))
				if err != nil {
					return nil, fmt.Errorf("parse workflow YAML: %w", err)
				}
				if err := Validate(w); err != nil {
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
