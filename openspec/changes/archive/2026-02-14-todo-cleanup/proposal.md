## Why

The codebase contains 4 TODO comments that are misleading, stale, or associated with dead code. Cleaning these up improves code clarity and prevents future developers from wasting time on already-resolved items.

## What Changes

- Remove unused `UpdateField` method from `ConfigState` (dead code — actual updates use `UpdateConfigFromForm` / `UpdateProviderFromForm`)
- Replace misleading `// TODO: Implement save logic` with accurate comment explaining save is handled by the caller
- Replace `// TODO: Implement ListModels proxying if needed` with descriptive comment of current behavior
- Extract `rpcTimeout` constant from hardcoded `30 * time.Second` values, removing `// TODO: Configurable timeout?` comment

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none — these are implementation-level cleanups with no spec-level behavior changes)

## Impact

- `internal/cli/onboard/state.go` — `UpdateField` method removed (no callers)
- `internal/cli/onboard/wizard.go` — comment clarified
- `internal/supervisor/proxy.go` — comment clarified
- `internal/supervisor/supervisor_test.go` — test comment updated
- `internal/security/rpc_provider.go` — `rpcTimeout` constant extracted, 3 call sites updated
