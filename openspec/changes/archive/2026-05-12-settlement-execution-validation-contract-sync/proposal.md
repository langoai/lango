## Why

The current settlement execution implementation treats an empty `transaction_receipt_id` as an actionable validation error before any denied business outcome is synthesized. Some architecture docs still present the deny reasons without that validation split, and the production-readiness spec still mentions an older `ExecuteRecommendation` method name instead of the current `Execute` entrypoint.

## What Changes

- Update the actual-settlement-execution architecture doc to distinguish request validation from denied execution outcomes.
- Update the production-readiness spec to reference `settlementexecution.Service.Execute` instead of the older `ExecuteRecommendation` name.
- Sync downstream docs/spec coverage for the settlement execution validation-vs-deny contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: settlement execution request-guard coverage now references the actual service entrypoint.
- `downstream-docs-sync`: settlement execution docs explain that empty transaction receipt ids fail validation before deny reasons apply.

## Impact

- Affected docs: `docs/architecture/actual-settlement-execution.md`
- Affected specs: `openspec/specs/production-readiness/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
