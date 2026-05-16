## Why

The settlement progression service still contains a panic path: escalating from an unknown current settlement progression status crashes instead of returning an actionable error. That is not production-safe for a policy-driven progression boundary.

## What Changes

- Replace the settlement-progression escalation panic with a structured error.
- Add regressions that lock the fail-closed contract for unknown current progression status.
- Sync `production-readiness` coverage for escalation fail-closed behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: settlement progression escalation now returns an actionable error instead of panicking on unknown current status.

## Impact

- Affected code: `internal/settlementprogression/service.go`, `internal/settlementprogression/types.go`
- Affected tests: `internal/settlementprogression/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
