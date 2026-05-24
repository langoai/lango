## Why

Several ontology governance and dynamic action tools declared required parameters in their schemas, but the wrapper layer still converted raw params with `fmt.Sprintf` and could leak `<nil>` into service execution instead of surfacing clear missing-parameter errors. That weakens the operator-facing contract for some of the highest-leverage ontology mutation paths.

## What Changes

- Tighten ontology governance and dynamic action tool wrappers so declared required params are extracted with `toolparam.RequireString`.
- Add regression coverage for missing governance/action parameters.
- Sync ontology and production-readiness coverage for the wrapper guard contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `ontology-tools`: governance and dynamic action wrappers now preserve actionable missing-parameter errors for declared inputs.
- `production-readiness`: wrapper-level request-guard coverage now includes the ontology governance/action cluster.

## Impact

- Affected code: `internal/ontology/tools.go`
- Affected tests: `internal/ontology/tools_test.go`
- Affected specs: `openspec/specs/ontology-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
