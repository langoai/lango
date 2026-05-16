## Why

The post-adjudication replay service already enforces fail-closed checks for missing receipt store and dispatcher dependencies, but those operator-facing contracts are not directly protected by regressions. That leaves a workflow-critical recovery boundary weakly verified.

## What Changes

- Add regression tests for missing receipt-store and dispatcher wiring in `postadjudicationreplay.Service`.
- Sync `production-readiness` coverage for replay dependency-guard behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: post-adjudication replay now has explicit regression coverage for missing-store and missing-dispatcher fail-closed behavior.

## Impact

- Affected tests: `internal/postadjudicationreplay/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
