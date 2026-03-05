## Context

The TUI settings editor (`internal/cli/settings/`) provides form-based configuration for all major features. Each feature follows a consistent pattern: a form constructor in `forms_*.go`, a menu entry in `menu.go`, a case in `editor.go`, and state binding in `tuicore/state_update.go`. MCP config (`config.MCPConfig`) already exists with 6 global fields but has no TUI form.

## Goals / Non-Goals

**Goals:**
- Expose MCP global settings (enabled, timeout, tokens, health check, reconnect) in TUI
- Follow existing form patterns exactly (NewCronForm, NewWorkflowForm)
- Place MCP in Infrastructure section alongside related automation features

**Non-Goals:**
- Individual MCP server management (add/remove/enable/disable) — already handled by CLI (`lango mcp add/remove/...`)
- MCP server-level config editing in TUI (transport, env, args)

## Decisions

1. **Form placement**: Infrastructure section, after Workflow Engine. MCP servers are infrastructure-level integrations, consistent with cron/background/workflow grouping.

2. **Field selection**: Only global MCPConfig fields (6 total). Server-specific fields are complex (transport, env vars, command args) and better served by CLI's `lango mcp add` interactive flow.

3. **Duration validation**: Use `time.ParseDuration` inline validation on timeout and interval fields, matching the pattern used in WorkflowForm's timeout field.

4. **Key prefix**: `mcp_` prefix for all field keys, consistent with `cron_`, `bg_`, `wf_` conventions.

## Risks / Trade-offs

- [Risk] Duration fields show "0s" when unconfigured → Acceptable; user sees default and can override.
- [Risk] No server list in TUI → Mitigated by CLI commands (`lango mcp list/add/remove`). TUI scope is intentionally limited to global settings.
