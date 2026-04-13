## 1. WebChat Playground

- [x] 1.1 Create `internal/gateway/playground/index.html` — self-contained HTML/CSS/JS with WebSocket JSON-RPC, markdown rendering, dark/light mode, connection status
- [x] 1.2 Create `internal/gateway/playground.go` — `go:embed` directive and `servePlayground` handler
- [x] 1.3 Add `GET /playground` route to the protected group in `server.go:setupRoutes()`

## 2. Session Isolation

- [x] 2.1 In `server.go:handleWebSocketConnection()`, assign `clientID` as session key when `SessionFromContext` returns empty

## 3. Auth Middleware Export

- [x] 3.1 Rename `requireAuth` → `RequireAuth` in `internal/gateway/middleware.go`
- [x] 3.2 Update `server.go:setupRoutes()` to use `RequireAuth`
- [x] 3.3 Update `middleware_test.go` to use `RequireAuth`

## 4. P2P Route Authentication

- [x] 4.1 Update `registerP2PRoutes` signature to accept `*gateway.AuthManager`
- [x] 4.2 Apply `gateway.RequireAuth(auth)` middleware to the `/api/p2p` route group
- [x] 4.3 Update `wirePostAgent` in `app.go` to pass `auth` parameter through to `registerP2PRoutes`

## 5. Approval Policy Warning

- [x] 5.1 Add WARN log in `app.go` when `approvalPolicy == "none"` during initialization

## 6. Documentation Drift Fixes

- [x] 6.1 Fix `docs/configuration.md` approval policy values: `always`→`all`, `dangerous`, `never`→`none`, add `configured`
- [x] 6.2 Fix `docs/configuration.md` metrics format: mark Prometheus as not yet implemented
- [x] 6.3 Update `docs/gateway/http-api.md`: add Playground endpoint, update P2P auth status
- [x] 6.4 Update `docs/getting-started/quickstart.md`: add playground tip after `lango serve`

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./internal/gateway/... ./internal/app/...` passes
- [x] 7.3 `go test ./...` passes (full suite)
