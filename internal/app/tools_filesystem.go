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
			Parameters: agent.Schema().
				Str("path", "The file path to read").
				Int("offset", "Start reading from this line number (1-indexed, default: read from beginning)").
				Int("limit", "Maximum number of lines to return (default: all lines)").
				Required("path").
				Build(),
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
			Parameters: agent.Schema().
				Str("path", "The directory path to list").
				Required("path").
				Build(),
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path := toolparam.OptionalString(params, "path", ".")
				return fsTool.ListDir(path)
			},
		},
		{
			Name:        "fs_write",
			Description: "Write content to a file",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: agent.Schema().
				Str("path", "The file path to write to").
				Str("content", "The content to write").
				Required("path", "content").
				Build(),
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
			Parameters: agent.Schema().
				Str("path", "The file path to edit").
				Int("startLine", "The starting line number (1-indexed)").
				Int("endLine", "The ending line number (inclusive)").
				Str("content", "The new content for the specified range").
				Required("path", "startLine", "endLine", "content").
				Build(),
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
			Parameters: agent.Schema().
				Str("path", "The directory path to create").
				Required("path").
				Build(),
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
			Parameters: agent.Schema().
				Str("path", "The path to delete").
				Required("path").
				Build(),
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
			Parameters: agent.Schema().
				Str("path", "The file path to inspect").
				Required("path").
				Build(),
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
