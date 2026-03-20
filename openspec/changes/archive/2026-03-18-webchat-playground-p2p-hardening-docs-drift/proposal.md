## Why

New users need a zero-setup way to experience the agent before configuring external channels. Currently the only interaction paths require Telegram/Discord/Slack setup or raw WebSocket tooling. Separately, P2P metadata endpoints expose node information without authentication when OIDC is active, and several documentation pages have drifted from the actual codebase (approval policy names, metrics format support).

## What Changes

- **WebChat Playground**: Embedded HTML/CSS/JS page served at `/playground` via `go:embed`, using the existing WebSocket JSON-RPC protocol for real-time streaming
- **Session Isolation**: Unauthenticated WebSocket clients receive a unique session key derived from their clientID, preventing cross-tab response bleed
- **P2P Route Authentication**: P2P metadata endpoints (`/api/p2p/*`) now use `RequireAuth` middleware — requires authentication when OIDC is configured, passes through in dev mode
- **`requireAuth` Export**: Renamed to `RequireAuth` so the `app` package can apply it to non-gateway routes
- **approvalPolicy "none" Warning**: Startup log warning when tool approval is completely disabled
- **Docs Factual Drift Fixes**: Corrected approval policy names (`all`/`dangerous`/`configured`/`none`), metrics format documentation, P2P auth status, and added playground references

## Capabilities

### New Capabilities
- `webchat-playground`: Embedded browser-based chat interface for testing agents without external channel setup

### Modified Capabilities
- `gateway-auth-middleware`: `requireAuth` renamed to `RequireAuth` (exported); unauthenticated clients now receive auto-assigned session keys for isolation
- `gateway-server`: New `/playground` route in the protected group
- `p2p-rest-api`: P2P metadata endpoints now require authentication when OIDC is configured
- `approval-policy`: Startup warning logged when policy is set to `"none"`

## Impact

- **Code**: `internal/gateway/` (new playground.go, playground/index.html; modified server.go, middleware.go), `internal/app/` (modified app.go, p2p_routes.go)
- **APIs**: New `GET /playground` endpoint; `/api/p2p/*` endpoints now auth-gated when OIDC active
- **Docs**: `docs/configuration.md`, `docs/security/tool-approval.md`, `docs/getting-started/quickstart.md`, `docs/gateway/http-api.md`
- **Tests**: `middleware_test.go` updated for `RequireAuth` rename
