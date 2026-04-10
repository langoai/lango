## Why

In Docker slim + `chromedp/headless-shell` sidecar environments, the `go-rod/rod` library panics instead of returning errors on Chrome CDP/WebSocket disconnection, crashing the entire process. Without panic recovery in the tool execution path and WebSocket goroutines, a single browser failure leads to complete service outage.

## What Changes

- Add `ErrBrowserPanic` sentinel error and `safeRodCall`/`safeRodCallValue` panic recovery wrappers in the browser tool core layer
- Wrap all rod/CDP method calls (Navigate, Screenshot, Click, Type, GetText, GetSnapshot, GetElementInfo, Eval, WaitForSelector, NewSession, Close) with panic recovery
- Add auto-reconnect logic in `SessionManager.EnsureSession()` on `ErrBrowserPanic` detection (close + retry once)
- Add `wrapBrowserHandler` in the application layer to catch panics and retry on `ErrBrowserPanic` at the tool handler level
- Add panic recovery to WebSocket `readPump`/`writePump` goroutines and isolate RPC handler panics so a single handler crash does not tear down the connection
- Add Chrome sidecar healthcheck and `service_healthy` dependency condition in docker-compose.yml

## Capabilities

### New Capabilities

### Modified Capabilities
- `tool-browser`: Add panic recovery layer around all rod/CDP calls and auto-reconnect on connection loss
- `docker-deployment`: Add Chrome sidecar healthcheck and healthy dependency condition
- `gateway-server`: Add panic recovery to WebSocket read/write pumps and RPC handler invocations

## Impact

- `internal/tools/browser/browser.go` — new error type, panic recovery wrappers, all methods wrapped
- `internal/tools/browser/session_manager.go` — auto-reconnect on ErrBrowserPanic
- `internal/app/tools.go` — wrapBrowserHandler applied to all browser tools
- `internal/gateway/server.go` — readPump/writePump/handleRPC panic recovery
- `docker-compose.yml` — Chrome service healthcheck
- `internal/tools/browser/panic_recovery_test.go` — new test file
