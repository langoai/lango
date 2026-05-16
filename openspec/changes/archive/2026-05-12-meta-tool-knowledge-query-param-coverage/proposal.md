## Why

The knowledge-artifact wrapper pass covered save/exportability/release approval, but the read-side knowledge tools still had one declaration drift: `search_knowledge` marked `query` as required while its handler accepted an empty query. `get_knowledge_history` also depended on a required `key` without direct regression coverage.

## What Changes

- Fix `search_knowledge` so its handler enforces the required `query` parameter.
- Add wrapper-level missing-parameter regressions for `get_knowledge_history` and `search_knowledge`.
- Extend meta-tools and production-readiness coverage for the knowledge read-tool wrapper guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: knowledge read tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request-guard coverage now includes the core knowledge history/search tool pair.

## Impact

- Affected code: `internal/app/tools_meta.go`
- Affected tests: `internal/app/tools_meta_validation_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
