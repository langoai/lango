## Why

The wrapper-level missing-parameter coverage is now strong across transaction and knowledge-artifact tools, but the learning/skill-management cluster still relied on implicit `toolparam` behavior. That left one real contract mismatch too: `search_learnings` declared `query` as required while its handler still accepted an empty query.

## What Changes

- Fix `search_learnings` so its handler enforces the required `query` parameter.
- Add wrapper-level missing-parameter regressions for `save_learning`, `search_learnings`, and `create_skill`.
- Extend meta-tools and production-readiness coverage for the learning/skill-management wrapper guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: the learning/skill-management tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request-guard coverage now includes the learning and skill management tool cluster.

## Impact

- Affected code: `internal/app/tools_meta.go`
- Affected tests: `internal/app/tools_meta_skills_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
