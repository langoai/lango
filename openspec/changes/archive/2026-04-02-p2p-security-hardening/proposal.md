## Why

External security review identified 5 fail-open behaviors in P2P tool execution paths. A remote peer connecting via libp2p can exploit these to: execute tools without container isolation (sandbox fallback), access files outside allowed directories (symlink escape), run arbitrary JavaScript in the browser (eval injection), invoke Dangerous-level tools like `exec` and `payment_send` (no SafetyLevel gate), and gain host process memory access (nil sandbox fallback).

## What Changes

- Sandbox executor gains `requireContainer` fail-closed mode — when true and no container runtime available, returns error instead of silently falling back to NativeRuntime
- P2P protocol handler refuses tool execution when `sandboxExec == nil` — returns `ErrNoSandboxExecutor` with `ResponseStatusDenied`
- Filesystem `validatePath()` calls `filepath.EvalSymlinks()` after `filepath.Abs()` to resolve symlinks before allowed/blocked path checks
- Filesystem `Delete()` accepts `context.Context` and restricts to `os.Remove` (single file) in P2P context instead of `os.RemoveAll`
- Browser tools validate URLs against private network blocklist (`localhost`, `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `file://`) in P2P context
- Browser `eval` action blocked for P2P requests
- New `SafetyLevelChecker` callback + `checkSafetyGate()` in P2P handler blocks tools above configured `maxSafetyLevel` (default: "moderate")
- Explicit `allowedTools` whitelist bypasses safety gate for specific tools
- `ctxkeys.WithP2PRequest`/`IsP2PRequest` propagates P2P origin through context chain

## Capabilities

### New Capabilities

- `p2p-sandbox-fail-closed`: Container runtime enforcement for P2P tool execution with configurable `requireContainer` flag
- `p2p-url-validation`: Private network URL blocking for browser tools in P2P context
- `p2p-safety-level-gate`: Tool safety level enforcement with configurable threshold and whitelist for P2P peers
- `fs-symlink-resolution`: Symlink-aware path validation preventing allowed-directory escape
- `fs-p2p-delete-restriction`: Single-file-only deletion for P2P-originated requests

### Modified Capabilities

- `p2p-protocol`: Handler now injects P2P context, checks sandbox availability, and enforces safety gate before tool execution
- `tool-filesystem`: `Delete` signature changed to accept `context.Context`
- `tool-browser`: Navigate and action handlers check P2P context for URL/eval restrictions

## Impact

- `internal/sandbox/container_executor.go` — fail-closed logic
- `internal/p2p/protocol/handler.go` — nil-safety, P2P context injection, safety gate
- `internal/p2p/protocol/messages.go` — `ErrNoSandboxExecutor`, `ErrToolSafetyBlocked` sentinel errors
- `internal/tools/filesystem/filesystem.go` — `EvalSymlinks`, P2P context helpers, Delete ctx
- `internal/tools/filesystem/tools.go` — pass ctx to Delete
- `internal/tools/browser/tools.go` — URL validation gate, eval block
- `internal/tools/browser/url_validator.go` — new file, private network CIDR checks
- `internal/ctxkeys/ctxkeys.go` — `WithP2PRequest`/`IsP2PRequest`
- `internal/config/types_p2p.go` — `RequireContainer`, `MaxSafetyLevel`, `AllowedTools` fields
- `internal/config/loader.go` — defaults: `RequireContainer: true`, `MaxSafetyLevel: "moderate"`
- `internal/agent/runtime.go` — `ParseSafetyLevel()`
- `internal/toolcatalog/catalog.go` — `GetToolSafetyLevel()`
