## Why

We already corrected stale public examples like `[y/N] y` to the actual shared prompt format `[y/N]: y`, but nothing prevents those low-signal regressions from reappearing in README or docs later.

## What Changes

- Add an executable repository test that rejects stale shared confirmation examples missing the colon separator
- Record the new docs-quality guard in the docs-only and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `docs-only`: shared confirmation example punctuation is regression-guarded
- `test-coverage`: public-doc prompt punctuation has an executable guard

## Impact

- Affected code: `internal/testutil/docs_quality_guard_test.go`
- Affected specs: `openspec/specs/docs-only/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
