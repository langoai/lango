# Design

## Root Cause

`LoadMCPFile` already returns errors for invalid files, but `MergedServers` and several MCP CLI commands ignore those errors. Missing optional scoped files and present invalid scoped files are therefore indistinguishable to callers.

## Approach

- Add strict scoped merge behavior that ignores `os.ErrNotExist` only.
- Include scope and path in errors returned for malformed, unreadable, or invalid scoped files.
- Update read-only MCP CLI commands to return merge errors directly.
- Update write MCP CLI commands to treat missing target files as empty config while preserving errors from present invalid target files.
- Keep any compatibility wrapper small and explicit if needed by existing callers.

## Error Policy

- Missing user or project config file: ignored.
- Present user or project config file that cannot be parsed or read: command/startup path receives an error containing scope and path.
- Existing target file for add/remove/enable/disable that is invalid: command fails and does not save replacement content.

## Downstream Impact

The behavior change is user-visible for MCP CLI commands. Public MCP docs need a short note explaining that missing scoped config files are optional, but invalid present files fail visibly.
