## Why

The RPC wallet already cleaned up pending address requests correctly, but the regression suite still treated address cleanup less symmetrically than the transaction/message signing paths. That leaves lifecycle hygiene weaker than it needs to be for long-lived interactive sessions.

## What Changes

- Add address cleanup regressions for companion-error and timeout exits.
- Keep runtime behavior unchanged; this is a coverage-hardening change.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: RPC wallet address-dispatch cleanup now has direct coverage for timeout and companion-error teardown.

## Impact

- Affected tests: `internal/wallet/rpc_wallet_test.go`
