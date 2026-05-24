## Why

The knowledge-runtime service already rejects a missing receipt store, but `SelectExecutionPath` still relies on downstream store behavior when its transaction receipt id is empty. That leaves an operator-facing request-validation contract weakly defined.

## What Changes

- Make `knowledgeruntime.Service.SelectExecutionPath` reject an empty transaction receipt id with an actionable validation error.
- Add regression coverage for that request guard.
- Sync `production-readiness` coverage for the knowledge-runtime request-id validation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: knowledge-runtime selection now has explicit validation coverage for missing transaction receipt ids.

## Impact

- Affected code: `internal/knowledgeruntime/service.go`
- Affected tests: `internal/knowledgeruntime/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
