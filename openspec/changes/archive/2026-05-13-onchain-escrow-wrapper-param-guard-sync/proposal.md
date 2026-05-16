## Why

The on-chain escrow tool cluster already declared `milestones` as required for `escrow_create` and `sellerPercent` as required for `escrow_resolve`, but the wrapper layer still let those values slip through to downstream logic. That weakens the operator-facing contract and makes parameter errors less actionable than the rest of the hardened tool surface.

## What Changes

- Tighten `escrow_create` to reject missing `milestones` at the wrapper layer.
- Tighten `escrow_resolve` to reject missing `sellerPercent` at the wrapper layer.
- Add regressions for both missing-parameter paths.
- Sync on-chain escrow and production-readiness coverage for the wrapper guard contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `onchain-escrow`: on-chain escrow tool wrappers now preserve actionable missing-parameter errors for required `milestones` and `sellerPercent`.
- `production-readiness`: wrapper-level request-guard coverage now includes the on-chain escrow tool cluster.

## Impact

- Affected code: `internal/app/tools_escrow.go`
- Affected tests: `internal/app/tools_escrow_test.go`
- Affected specs: `openspec/specs/onchain-escrow/spec.md`, `openspec/specs/production-readiness/spec.md`
