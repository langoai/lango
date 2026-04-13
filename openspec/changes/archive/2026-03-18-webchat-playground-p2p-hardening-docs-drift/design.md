## Context

The gateway server already supports WebSocket JSON-RPC for real-time agent interaction, with session-scoped event broadcasting, OIDC authentication middleware, and structured error events. The playground builds entirely on these existing primitives — no new protocols or agent-side changes are needed.

P2P metadata endpoints (`/api/p2p/*`) were originally public because they expose "only metadata." However, when OIDC is active, this creates an inconsistency — all other non-health endpoints require authentication.

Several documentation pages contain factual drift from code: approval policy names use old terminology, metrics format claims Prometheus support that was never implemented.

## Goals / Non-Goals

**Goals:**
- Provide a zero-setup browser UI for testing agents via the existing WebSocket protocol
- Isolate unauthenticated WebSocket clients so two browser tabs don't see each other's responses
- Protect P2P metadata endpoints with the same auth middleware used by `/ws` and `/status`
- Fix documented approval policy names and metrics format to match actual code
- Add playground references to quickstart and HTTP API docs

**Non-Goals:**
- Full-featured chat UI (conversation history, themes, settings persistence)
- P2P response schema changes for unauthenticated mode (separate design needed)
- Prometheus metrics format implementation
- General sandbox isolation redesign (P2P-specific → general)

## Decisions

### Decision 1: Playground as `go:embed` single HTML file

Embed a self-contained HTML/CSS/JS file using Go's `go:embed` directive. No external CDN dependencies, no build toolchain, no static file serving configuration.

**Rationale**: Minimizes deployment complexity. The playground is a debugging/testing tool, not a production UI. A single embedded file keeps the binary self-contained.

**Alternative**: Serve from filesystem or use a frontend framework. Rejected — adds build/deploy complexity for a developer tool.

### Decision 2: Session isolation via clientID assignment

When `SessionFromContext()` returns empty (no OIDC), assign the already-unique `clientID` (e.g., `"ui-{UnixNano}"`) as the session key. This makes `BroadcastToSession()` filtering work automatically.

**Rationale**: `clientID` is already generated as unique (line 668 in server.go). `handleChatMessage` uses `client.SessionKey` when non-empty. No new state management needed.

**Alternative**: Generate a separate session UUID. Rejected — unnecessary; clientID already serves this purpose.

### Decision 3: `/playground` in the protected route group

Place `/playground` alongside `/status` in the `RequireAuth` middleware group.

**Rationale**: When OIDC is active, if `/playground` were public but `/ws` required auth, users would see the page but the WebSocket would fail with 401 — a confusing half-working state. When OIDC is not configured, `RequireAuth(nil)` passes through, so the playground remains accessible.

### Decision 4: Export `requireAuth` → `RequireAuth`

Rename the middleware constructor to be exported so `internal/app` can apply it to P2P routes without duplicating logic.

**Rationale**: The `app` package registers P2P routes on the gateway router but couldn't use the auth middleware because it was unexported. Exporting follows Go convention for cross-package use.

### Decision 5: WebSocket URL via `location` object

```javascript
const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
const url = proto + '//' + location.host + '/ws';
```

**Rationale**: Works behind reverse proxies, non-default ports, and HTTPS. Hardcoding `ws://localhost:18789/ws` would break in any non-default deployment.

## Risks / Trade-offs

- **[Risk]** Playground HTML grows stale vs. evolving WS protocol → **Mitigation**: Playground only uses stable JSON-RPC events already documented in websocket.md. Protocol changes naturally break it, forcing updates.
- **[Risk]** `RequireAuth` export is a public API change → **Mitigation**: Only used by `internal/` packages. Not part of external API surface.
- **[Risk]** Session isolation changes broadcast behavior for unauthenticated clients → **Mitigation**: Previous behavior (broadcast to all UI clients) was a bug, not a feature. Isolated sessions are strictly more correct.
