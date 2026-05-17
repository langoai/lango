## Context

Background task management already exists in two forms:

- in-process manager methods: `List`, `Status`, `Result`, `Cancel`
- agent tools: `bg_list`, `bg_status`, `bg_result`, `bg_cancel`

The missing piece is a gateway REST surface that lets root CLI commands reach the running server process. Existing gateway-backed CLIs use `--addr`, JSON responses, bounded HTTP clients, and table/json output patterns.

## Decisions

### Gateway API

Add `/api/bg` routes on the app gateway router:

- `GET /api/bg/tasks` returns `{ "tasks": [...] }`
- `GET /api/bg/tasks/{id}` returns `{ "task": ... }`
- `GET /api/bg/tasks/{id}/result` returns `{ "result": "..." }`
- `POST /api/bg/tasks/{id}/cancel` returns `{ "id": "...", "cancelled": true }`

All routes use `gateway.RequireAuth(auth)` so local unauthenticated development remains easy when auth is nil, while authenticated deployments protect prompts/results/cancel mutation.

### DTO Boundary

Do not expose `background.TaskSnapshot` directly as the public API. Define a route-local DTO with stable string fields:

- `id`
- `status`
- `prompt`
- `originChannel`
- `originSession`
- `startedAt`
- `completedAt`
- `duration`
- `error`
- `result`

This avoids leaking the internal enum integer while still preserving useful task details.

### CLI Boundary

Refactor `internal/cli/bg` from `func() (*background.Manager, error)` to a small `Client` interface:

- `List(ctx) ([]Task, error)`
- `Status(ctx, id) (Task, error)`
- `Result(ctx, id) (string, error)`
- `Cancel(ctx, id) error`

Provide:

- `NewInProcessClient(managerProvider)` for embedded tests/callers
- `NewGatewayClient(addr, httpClient)` for root CLI

The root command resolves `--addr` first, then falls back to `boot.Config.Server.Host/Port`.

### Output

Keep current table/plain outputs for compatibility. Add `--output table|json` to the command group so remote results can be scripted. All output continues through Cobra command writers.

## Risks

- A running server with auth enabled will reject unauthenticated CLI calls unless session/cookie support is added later. This change should surface clear gateway errors rather than pretending to bypass auth.
- Background state remains in-memory. The CLI must not imply persistence across server restarts.
- Result and prompt text may be sensitive; route auth must be consistently applied.
