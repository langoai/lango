## Why

The librarian inquiry tools already declare `inquiry_id` as a required wrapper input for dismissal, and the handler rejects missing values. But there was no direct tool-entrypoint regression proving that failure happens before inquiry lookup, and the public docs did not state that contract explicitly.

## What Changes

- Add direct tool-entrypoint regression coverage for missing `inquiry_id` on `librarian_dismiss_inquiry`.
- Update README and multi-agent docs to describe the librarian inquiry required-input contract.
- Sync proactive-librarian, downstream-docs, and production-readiness specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `proactive-librarian`: inquiry dismissal required-input guard is now directly covered.
- `downstream-docs-sync`: librarian inquiry docs now mention the required wrapper input.
- `production-readiness`: wrapper guard coverage now includes librarian inquiry tools.

## Impact

- Affected tests: `internal/librarian/tools_test.go`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/proactive-librarian/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
