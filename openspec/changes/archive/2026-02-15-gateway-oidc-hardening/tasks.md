## 1. Auth Middleware (middleware.go — NEW)

- [x] 1.1 Create `internal/gateway/middleware.go` with unexported `contextKey` type and `sessionContextKey` constant
- [x] 1.2 Implement `SessionFromContext(ctx)` to extract session key from context
- [x] 1.3 Implement `requireAuth(auth *AuthManager)` chi middleware — nil auth pass-through, cookie validation, context binding
- [x] 1.4 Implement `makeOriginChecker(allowedOrigins []string)` — nil/empty returns nil, `*` allows all, specific list whitelist with trailing slash normalization
- [x] 1.5 Implement `isSecure(r *http.Request)` — checks `r.TLS` and `X-Forwarded-Proto` header

## 2. Server Hardening (server.go — MODIFY)

- [x] 2.1 Add `AllowedOrigins []string` to `Config` struct
- [x] 2.2 Change `RPCHandler` type signature to `func(client *Client, params json.RawMessage) (interface{}, error)`
- [x] 2.3 Update `New()` to use `makeOriginChecker(cfg.AllowedOrigins)` for WebSocket upgrader
- [x] 2.4 Update `setupRoutes()` — public (`/health`), auth routes, protected group (`/ws`, `/status`) with `requireAuth`, separate `/companion`
- [x] 2.5 Update `handleWebSocketConnection()` — extract session via `SessionFromContext(r.Context())` and bind to `Client.SessionKey`
- [x] 2.6 Update `handleChatMessage()` — use `client.SessionKey` when authenticated, fallback to request key or "default"
- [x] 2.7 Add nil agent check in `handleChatMessage()`
- [x] 2.8 Update all RPC handler signatures to accept `client *Client` parameter
- [x] 2.9 Update `readPump()` to pass `c` (client) to handler calls

## 3. Config Update (types.go — MODIFY)

- [x] 3.1 Add `AllowedOrigins []string` field to `ServerConfig` with mapstructure/json tags

## 4. Wiring Update (wiring.go — MODIFY)

- [x] 4.1 Pass `cfg.Server.AllowedOrigins` to `gateway.Config` in `initGateway()`

## 5. Auth Improvements (auth.go — MODIFY)

- [x] 5.1 Add `middleware.Throttle(10)` to auth route group in `RegisterRoutes()`
- [x] 5.2 Change state cookie name from `oauth_state` to `oauth_state_{provider}` in `handleLogin()`
- [x] 5.3 Update `handleCallback()` — use per-provider state cookie name, delete state cookie after validation
- [x] 5.4 Replace `r.TLS != nil` with `isSecure(r)` for all cookie Secure flags
- [x] 5.5 Change callback response from plaintext email to JSON `{"status":"authenticated","sessionKey":"..."}`
- [x] 5.6 Implement `handleLogout()` — delete session from store, clear cookie, return JSON response
- [x] 5.7 Register `POST /auth/logout` route

## 6. Tests

- [x] 6.1 Update `server_test.go` — fix RPCHandler signature in existing `TestGatewayServer`
- [x] 6.2 Add `TestChatMessage_UnauthenticatedUsesDefault` and `TestChatMessage_AuthenticatedUsesOwnSession`
- [x] 6.3 Create `middleware_test.go` with `mockStore` implementation
- [x] 6.4 Add `TestRequireAuth_NilAuthPassesThrough`, `TestRequireAuth_NoCookieReturns401`, `TestRequireAuth_InvalidSessionReturns401`, `TestRequireAuth_ValidSessionSetsContext`
- [x] 6.5 Add `TestSessionFromContext_Empty`
- [x] 6.6 Add `TestMakeOriginChecker_EmptyReturnsNil`, `TestMakeOriginChecker_WildcardAllowsAll`, `TestMakeOriginChecker_SpecificOriginsMatch`, `TestMakeOriginChecker_TrailingSlashNormalized`
- [x] 6.7 Add `TestIsSecure_DirectTLS`, `TestIsSecure_XForwardedProto`
- [x] 6.8 Add `TestLogout_ClearsSessionAndCookie`
- [x] 6.9 Add `TestStateCookie_PerProviderName`

## 7. Verification

- [x] 7.1 Run `go build ./...` — zero errors
- [x] 7.2 Run `go vet ./...` — zero warnings
- [x] 7.3 Run `go test ./internal/gateway/... -v` — all tests pass
- [x] 7.4 Run `go test ./internal/config/... -v` — config tests pass
