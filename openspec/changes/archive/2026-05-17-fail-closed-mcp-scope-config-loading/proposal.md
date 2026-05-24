# Fail Closed on MCP Scoped Config Load Errors

## Why

MCP server configuration is merged from profile, user, and project scopes. The previous merge path ignored user and project file load errors, so malformed or unreadable scoped config could be silently skipped. CLI commands could then report "No MCP servers configured" or "server not found", and write commands could overwrite a malformed file.

## What Changes

- Preserve optional behavior for missing user and project MCP config files.
- Return actionable errors for present-but-invalid scoped MCP config files.
- Prevent MCP CLI write commands from overwriting invalid existing scoped config.
- Keep the change scoped to MCP config loading, CLI callers, docs, and tests.

## Non-Goals

- Redesign MCP configuration storage.
- Add new MCP transports or server lifecycle behavior.
- Change remote MCP connection failure behavior after valid config is loaded.
