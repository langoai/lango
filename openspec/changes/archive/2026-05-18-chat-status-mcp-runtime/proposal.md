## Why

The interactive chat `/status` command currently classifies MCP as "configured but not active in TUI mode" whenever `mcp.enabled` is true. That is stale for local chat and cockpit paths because the app can build the extension module and initialize an MCP manager before the TUI starts.

## What Changes

- Pass real runtime feature state into the chat model.
- Render MCP as active when the local chat/cockpit app has an MCP manager.
- Keep non-local lifecycle features such as gateway, cron, P2P, and payment marked as configured but inactive in local TUI mode.
- Add tests that prevent `/status` from reporting MCP inactive when runtime state says it is active.

## Impact

- TUI status output becomes less misleading.
- No MCP initialization behavior or server lifecycle behavior changes.
- Public docs for TUI mode are updated to state that `/status` reflects runtime-initialized MCP.
