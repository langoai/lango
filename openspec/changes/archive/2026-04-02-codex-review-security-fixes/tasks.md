## 1. Filesystem P2P Context Fix

- [x] 1.1 Replace `IsP2PContext(ctx)` with `ctxkeys.IsP2PRequest(ctx)` in `filesystem.go` Delete method
- [x] 1.2 Remove unused `ctxKeyP2P`, `WithP2PContext`, `IsP2PContext` from `filesystem.go`
- [x] 1.3 Add `ctxkeys` import to `filesystem.go`

## 2. Safety Gate Wiring

- [x] 2.1 Add `SetSafetyGate()` call in `app.go` `wirePostAgent()` after approval setup
- [x] 2.2 Check `ParseSafetyLevel` boolean return, fallback to `SafetyLevelModerate` on invalid
- [x] 2.3 Add `MaxSafetyLevel: "moderate"` default to `DefaultConfig()` in `loader.go`

## 3. Browser SSRF Hardening

- [x] 3.1 Add DNS resolution (`net.LookupIP`) to `ValidateURLForP2P` in `url_validator.go`
- [x] 3.2 Extract `checkIPPrivate` helper for reuse between IP literal and DNS resolved checks
- [x] 3.3 Add `CurrentURL(sessionID)` method to browser `Tool` in `browser.go`
- [x] 3.4 Add post-navigation URL re-validation in `browser_navigate` handler in `tools.go`

## 4. Container Sandbox Fail-Closed

- [x] 4.1 Check `RequireContainer` before subprocess fallback in `app.go`
- [x] 4.2 Add nil guard on `sbxExec` before calling `SetSandboxExecutor`

## 5. Paid Tool Safety Ordering

- [x] 5.1 Move `checkSafetyGate` before `payGate.Check` in `handleToolInvokePaid`
- [x] 5.2 Update step-number comments in handler flow

## 6. Docs & Observability

- [x] 6.1 Update ADK version references to v0.6.0 in `index.md`, `overview.md`, `project-structure.md`
- [x] 6.2 Fix `evictOldestSession` to skip eviction when `MaxSessions <= 0`

## 7. Verification

- [x] 7.1 Run `go build ./...` — passes
- [x] 7.2 Run tests for all affected packages — all pass
