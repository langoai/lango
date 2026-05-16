## Why

The knowledge-runtime service currently assumes its receipt store is always present. If that dependency is missing, opening a transaction or selecting an execution path can panic instead of returning an actionable error. That weakens a core workflow boundary.

## What Changes

- Make `knowledgeruntime.Service` fail closed when its receipt store is unavailable.
- Add regressions for missing-store behavior across the service entrypoints.
- Sync `production-readiness` coverage for knowledge-runtime store dependency guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: knowledge-runtime entrypoints now return actionable unavailable-store errors instead of panicking.

## Impact

- Affected code: `internal/knowledgeruntime/service.go`
- Affected tests: `internal/knowledgeruntime/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
