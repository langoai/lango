## Why

The code and tests already treat `gvisor` as a stub runtime that always reports unavailable, but the public configuration tables still read like `gvisor` is a normal ready-to-use backend. That mismatch weakens production trust and makes runtime selection errors feel arbitrary.

## What Changes

- Add a regression that checks the explicit `gvisor` runtime path returns actionable wording, not just `ErrRuntimeUnavailable`.
- Update public configuration tables to state that `gvisor` is currently a stub.
- Sync production-readiness and downstream-docs specs with that public contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: explicit `gvisor` runtime requests now have a direct actionable-error regression.
- `downstream-docs-sync`: runtime tables document the current gVisor stub state honestly.

## Impact

- Affected code: `internal/sandbox/container_executor_test.go`
- Affected docs: `README.md`, `docs/configuration.md`
- Affected specs: `openspec/specs/production-readiness/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
