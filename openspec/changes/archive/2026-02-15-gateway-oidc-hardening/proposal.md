## Why

The Gateway OIDC authentication system has critical security vulnerabilities: WebSocket/HTTP endpoints lack authentication enforcement, session fixation allows arbitrary session access, and WebSocket CORS is fully open. While OIDC login/callback flows are correctly implemented, the authentication results are never enforced on protected endpoints, rendering the entire auth system ineffective.

## What Changes

- Add `requireAuth` middleware that enforces `lango_session` cookie validation on protected routes (`/ws`, `/status`), with pass-through when auth is not configured (dev mode)
- **BREAKING**: Change `RPCHandler` signature from `func(params json.RawMessage)` to `func(client *Client, params json.RawMessage)` to bind authenticated sessions to WebSocket clients
- Add configurable CORS origin checking for WebSocket connections via `AllowedOrigins` config
- Bind authenticated session keys to WebSocket clients, preventing session fixation in `handleChatMessage`
- Add `POST /auth/logout` endpoint for explicit session invalidation
- Fix state cookie collision by using per-provider names (`oauth_state_{provider}`)
- Apply `isSecure()` helper for reverse-proxy-aware cookie Secure flags
- Add rate limiting (`Throttle(10)`) on auth endpoints
- Remove plaintext email from login response, return structured JSON instead

## Capabilities

### New Capabilities
- `gateway-auth-middleware`: Authentication middleware for route protection, session-context binding, and CORS origin checking

### Modified Capabilities
- `oidc-auth`: Add logout endpoint, per-provider state cookies, rate limiting, secure cookie handling, and structured JSON responses
- `gateway-server`: Route protection with auth middleware, RPCHandler signature change, AllowedOrigins config, session-bound WebSocket clients

## Impact

- **Code**: `internal/gateway/` (new middleware.go, modified server.go, auth.go), `internal/config/types.go`, `internal/app/wiring.go`
- **APIs**: RPCHandler signature is a breaking change for any external handler registrations
- **Config**: New `server.allowedOrigins` field in ServerConfig
- **Dependencies**: No new dependencies (uses existing chi middleware and gorilla/websocket)
