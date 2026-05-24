## Design

Add a small `RuntimeFeatures` value to the chat model dependencies. It is a UI-facing snapshot, not a direct dependency on the app package or MCP manager type.

`cmdStatus` uses configuration for configured feature intent and `RuntimeFeatures` for features that may actually initialize in local chat mode. For this change, only MCP needs runtime truth because gateway, cron, P2P, and payment remain outside the local-chat lifecycle.

## Data Flow

1. `app.New(..., app.WithLocalChat())` may populate `application.MCPManager`.
2. CLI entrypoints convert that into `chat.RuntimeFeatures{MCPActive: application.MCPManager != nil}`.
3. Cockpit/workbench deps carry the same runtime snapshot to their embedded chat composer.
4. `/status` renders MCP as active when `MCPActive` is true.

## Non-Goals

- Do not start MCP if it is not already initialized.
- Do not change lifecycle priority behavior.
- Do not treat gateway, cron, P2P, or payment as active in local chat mode.
