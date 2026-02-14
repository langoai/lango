## Context

The Gateway OIDC authentication system correctly implements login/callback flows but fails to enforce authentication on protected endpoints. A security audit identified CRITICAL (3), HIGH (3), and MEDIUM (3) vulnerabilities. The existing multi-layered defense (SecretScanner, filesystem blocking, env whitelist) protects against agent-based secret exfiltration, but the gateway itself allows unauthenticated access to WebSocket and HTTP endpoints.

Current architecture:
- OIDC login → session cookie issued → endpoints ignore the cookie
- WebSocket `CheckOrigin` returns `true` for all origins
- `handleChatMessage` accepts arbitrary `sessionKey` from client input
- No logout mechanism exists

## Goals / Non-Goals

**Goals:**
- Enforce session-based authentication on `/ws` and `/status` when OIDC is configured
- Eliminate session fixation by binding authenticated sessions to WebSocket clients
- Restrict WebSocket CORS to configured origins
- Provide logout capability and prevent state cookie collision across providers
- Apply secure cookie flags behind reverse proxies
- Rate-limit authentication endpoints

**Non-Goals:**
- Token-based auth (JWT bearer) — cookie-based sessions are sufficient for the gateway
- Per-route authorization (role-based access control) — all authenticated users have equal access
- Companion endpoint OIDC auth — companions use device-based auth with origin restriction only
- Breaking backward compatibility for unauthenticated deployments — auth nil = pass-through

## Decisions

### D1: Auth middleware as chi route group middleware
**Decision**: Implement `requireAuth` as a standard chi middleware applied via `r.Group()`.
**Rationale**: Chi's route groups naturally segment public/protected routes. The middleware checks `lango_session` cookie → validates against session store → sets session key in request context. When `auth == nil` (no OIDC configured), the middleware is a no-op pass-through, preserving backward compatibility for local/dev deployments.
**Alternative considered**: Per-handler auth checks — rejected because it's error-prone and requires every new handler to remember to check auth.

### D2: Session binding via context propagation
**Decision**: `requireAuth` middleware stores the validated session key in `context.Value`. `handleWebSocketConnection` extracts it via `SessionFromContext(ctx)` and binds it to `Client.SessionKey`. `handleChatMessage` uses `Client.SessionKey` when non-empty, ignoring any client-provided `sessionKey`.
**Rationale**: This prevents session fixation — authenticated clients cannot specify arbitrary session keys. Unauthenticated clients (auth disabled) retain the old behavior of client-specified or "default" keys.
**Alternative considered**: First-message WebSocket auth — rejected as primary method because HTTP cookie validation during upgrade is simpler and more secure. First-message can be added later for API client support.

### D3: Origin checking via configurable allowlist
**Decision**: New `AllowedOrigins []string` config field. `makeOriginChecker` builds a gorilla `CheckOrigin` function: empty list → `nil` (gorilla default = same-origin), `["*"]` → allow all, specific list → whitelist match.
**Rationale**: Flexible for development (`*`), production (specific origins), and default (same-origin). Trailing slashes are normalized.

### D4: RPCHandler signature change
**Decision**: Change `RPCHandler` from `func(params json.RawMessage)` to `func(client *Client, params json.RawMessage)`.
**Rationale**: Handlers need access to the calling client's session context. This is a mechanical change — all existing handlers add the `client` or `_ *Client` parameter. The `readPump` method passes `c` (the reading client) to handlers.

### D5: Per-provider state cookies
**Decision**: State cookie name changes from `oauth_state` to `oauth_state_{provider}`. State cookie is deleted after successful callback.
**Rationale**: Prevents state overwrite when initiating concurrent login flows with multiple providers.

## Risks / Trade-offs

- **[RPCHandler breaking change]** → All handler call sites must update. Mitigated by mechanical nature of the change (add `client` parameter). No external consumers are known.
- **[Reverse proxy trust]** → `isSecure()` trusts `X-Forwarded-Proto` header. If deployed without a trusted proxy, an attacker could set this header. Mitigated by standard deployment behind nginx/caddy which strips/overrides this header.
- **[Rate limiting granularity]** → `Throttle(10)` limits concurrent requests server-wide, not per-IP. Sufficient for current scale; per-IP throttling can be added via middleware upgrade if needed.
- **[Cookie-only WebSocket auth]** → API clients that cannot send cookies must rely on auth-disabled mode. Future enhancement could add first-message token auth for programmatic access.
