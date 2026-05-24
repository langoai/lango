## Why

The learning and skill-management wrapper pass covered `save_learning`, `search_learnings`, and `create_skill`, but the remaining skill read/import entrypoints still relied on implicit parameter validation without direct regressions. `view_skill` and `import_skill` are also operator-facing meta tools and should keep explicit first-guard contracts.

## What Changes

- Add wrapper-level missing-parameter regressions for `view_skill` and `import_skill`.
- Extend meta-tools and production-readiness coverage for the remaining skill read/import wrapper guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: skill read/import tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request-guard coverage now includes the remaining skill-management read/import entrypoints.

## Impact

- Affected tests: `internal/app/tools_meta_skills_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
