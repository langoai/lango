package app

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/tools/filesystem"
)

func buildFilesystemTools(fsTool *filesystem.Tool) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "fs_read",
			Description: "Read a file. Supports optional offset/limit for partial reads.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":   map[string]interface{}{"type": "string", "description": "The file path to read"},
					"offset": map[string]interface{}{"type": "integer", "description": "Start reading from this line number (1-indexed, default: read from beginning)"},
					"limit":  map[string]interface{}{"type": "integer", "description": "Maximum number of lines to return (default: all lines)"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}

				offset := toolparam.OptionalInt(params, "offset", 0)
				limit := toolparam.OptionalInt(params, "limit", 0)

				if offset > 0 || limit > 0 {
					return fsTool.ReadWithMeta(path, offset, limit)
				}
				return fsTool.Read(path)
			},
		},
		{
			Name:        "fs_list",
			Description: "List contents of a directory",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The directory path to list"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path := toolparam.OptionalString(params, "path", ".")
				return fsTool.ListDir(path)
			},
		},
		{
			Name:        "fs_write",
			Description: "Write content to a file",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "The file path to write to"},
					"content": map[string]interface{}{"type": "string", "description": "The content to write"},
				},
				"required": []string{"path", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}
				content := toolparam.OptionalString(params, "content", "")
				return nil, fsTool.Write(path, content)
			},
		},
		{
			Name:        "fs_edit",
			Description: "Edit a file by replacing a line range",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":      map[string]interface{}{"type": "string", "description": "The file path to edit"},
					"startLine": map[string]interface{}{"type": "integer", "description": "The starting line number (1-indexed)"},
					"endLine":   map[string]interface{}{"type": "integer", "description": "The ending line number (inclusive)"},
					"content":   map[string]interface{}{"type": "string", "description": "The new content for the specified range"},
				},
				"required": []string{"path", "startLine", "endLine", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}
				content := toolparam.OptionalString(params, "content", "")
				startLine := toolparam.OptionalInt(params, "startLine", 0)
				endLine := toolparam.OptionalInt(params, "endLine", 0)
				return nil, fsTool.Edit(path, startLine, endLine, content)
			},
		},
		{
			Name:        "fs_mkdir",
			Description: "Create a directory",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The directory path to create"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}
				return nil, fsTool.Mkdir(path)
			},
		},
		{
			Name:        "fs_delete",
			Description: "Delete a file or directory",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The path to delete"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}
				return nil, fsTool.Delete(path)
			},
		},
		{
			Name:        "fs_stat",
			Description: "Get file metadata (size, line count, modification time) without reading content",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The file path to inspect"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, err := toolparam.RequireString(params, "path")
				if err != nil {
					return nil, err
				}
				return fsTool.Stat(path)
			},
		},
	}
}
