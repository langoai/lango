## Why

The RPC wallet address path cleanup contract is now stronger, but the quality goal here is to keep closing the last obvious lifecycle gaps with explicit regression tests rather than relying on implementation symmetry by inspection.

## What Changes

- Lock in address pending cleanup on the last two obvious non-success exits: timeout and companion error.
- Keep runtime code unchanged; this is a test-only hardening change.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: RPC wallet address lifecycle cleanup now has direct timeout and companion-error regression coverage.

## Impact

- Affected tests: `internal/wallet/rpc_wallet_test.go`
