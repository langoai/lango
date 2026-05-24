## Why

The economy-layer escrow tools declared several required parameters in their schemas, and the operator docs already describe those parameters as required. But parts of the wrapper layer still accepted missing values and only failed later through downstream escrow engine behavior, which weakens the operator-facing contract and obscures the real input problem.

## What Changes

- Tighten economy escrow tool wrappers to enforce their declared required parameters directly.
- Add regression coverage for the new missing-parameter paths.
- Sync economy-escrow and production-readiness coverage for the wrapper guard contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `economy-escrow`: economy-layer escrow tools now preserve actionable wrapper-level missing-parameter errors for declared required inputs.
- `production-readiness`: wrapper-level request-guard coverage now includes the economy escrow tool cluster.

## Impact

- Affected code: `internal/economy/tools.go`
- Affected tests: `internal/economy/tools_test.go`
- Affected specs: `openspec/specs/economy-escrow/spec.md`, `openspec/specs/production-readiness/spec.md`
