## Why

The RPC wallet already cleaned pending request state through deferred teardown, but the regression suite still verified that guarantee unevenly across request types and failure modes. The implementation should not rely on that asymmetry staying benign.

## What Changes

- Add missing cleanup regressions for sign-transaction success, companion error, sender error, and context cancellation.
- Add missing cleanup regressions for sign-message success, timeout, companion error, sender error, and context cancellation.
- Keep runtime code unchanged; this is a test-coverage hardening pass.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: RPC wallet request lifecycle cleanup is now covered symmetrically across success and failure paths.

## Impact

- Affected tests: `internal/wallet/rpc_wallet_test.go`
