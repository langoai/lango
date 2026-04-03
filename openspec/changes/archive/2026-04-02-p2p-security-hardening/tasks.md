# P2P Security Hardening — Tasks

## Implementation

- [x] 1.1 Add `requireContainer` field to `ContainerExecutor` with fail-closed logic
- [x] 1.2 Add `ErrNoSandboxExecutor` sentinel error to `messages.go`
- [x] 1.3 Refuse tool execution in handler when `sandboxExec == nil`
- [x] 1.4 Add `RequireContainer` config field to `types_p2p.go`
- [x] 1.5 Set `RequireContainer: true` default in `loader.go`
- [x] 2.1 Add `filepath.EvalSymlinks()` to `validatePath` in `filesystem.go`
- [x] 2.2 Add `WithP2PContext`/`IsP2PContext` helpers to `filesystem.go`
- [x] 2.3 Change `Delete` signature to accept `context.Context`
- [x] 2.4 Restrict `RemoveAll` to `Remove` in P2P context
- [x] 2.5 Update `fs_delete` handler in `tools.go` to pass ctx
- [x] 3.1 Create `url_validator.go` with `ValidateURLForP2P` + private network CIDR checks
- [x] 3.2 Add `ErrBlockedURL` and `ErrEvalBlockedP2P` sentinel errors
- [x] 3.3 Add `WithP2PRequest`/`IsP2PRequest` to `ctxkeys.go`
- [x] 3.4 Apply `WithP2PRequest(ctx)` in handler `handleToolInvoke` and `handleToolInvokePaid`
- [x] 3.5 Add URL validation gate in `browser_navigate` handler
- [x] 3.6 Block `eval` action in `browser_action` handler for P2P context
- [x] 4.1 Add `ParseSafetyLevel(string)` to `agent/runtime.go`
- [x] 4.2 Add `GetToolSafetyLevel(name)` to `toolcatalog/catalog.go`
- [x] 4.3 Add `SafetyLevelChecker` type and `SetSafetyGate` method to handler
- [x] 4.4 Add `checkSafetyGate` method with whitelist bypass
- [x] 4.5 Add safety gate checks in both invoke paths
- [x] 4.6 Add `MaxSafetyLevel` and `AllowedTools` config fields
- [x] 4.7 Add `ErrToolSafetyBlocked` sentinel error

## Testing

- [x] 5.1 Container executor fail-closed test (2 table cases)
- [x] 5.2 Handler nil sandbox executor test (free + paid paths)
- [x] 5.3 Filesystem symlink escape test (4 cases)
- [x] 5.4 Filesystem P2P delete restriction test (4 cases)
- [x] 5.5 Browser URL validator tests (internal blocked, external allowed)
- [x] 5.6 Browser eval P2P blocking tests
- [x] 5.7 ParseSafetyLevel tests (valid/invalid inputs)
- [x] 5.8 GetToolSafetyLevel tests (known/unknown tools)
- [x] 5.9 Safety gate handler tests (block dangerous, whitelist bypass, backward compat)
- [x] 5.10 Messages sentinel error test update
