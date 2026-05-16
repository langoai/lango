## Why

We manually cleaned archive-generated placeholder `Purpose` text from main specs, but nothing currently prevents the same low-signal text from being reintroduced later. That makes spec hygiene dependent on memory and review luck.

## What Changes

- Add an executable repository test that rejects archived-purpose placeholder text in main specs
- Record the new regression guard in the docs-only and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `docs-only`: placeholder purpose hygiene is now regression-guarded
- `test-coverage`: main-spec hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/spec_quality_guard_test.go`
- Affected specs: `openspec/specs/docs-only/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
