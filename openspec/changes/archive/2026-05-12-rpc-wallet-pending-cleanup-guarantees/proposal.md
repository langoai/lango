## Why

The RPC wallet already handles request/response, timeout, and sender-error paths, but stale pending-request entries are the kind of lifecycle bug that silently accumulates in long-lived sessions. That cleanup guarantee deserves direct regression coverage.

## What Changes

- Add wallet regressions that verify pending request maps are cleaned up after address sender errors.
- Add wallet regressions that verify pending request maps are cleaned up after sign-message sender errors.
- Add wallet regressions that verify pending request maps are cleaned up after address context cancellation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: RPC wallet dispatch coverage now includes pending-state cleanup guarantees on non-success paths.

## Impact

- Affected tests: `internal/wallet/rpc_wallet_test.go`
