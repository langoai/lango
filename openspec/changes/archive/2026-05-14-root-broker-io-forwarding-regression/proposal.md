## Why

The `runMain()` broker-mode path is already seam-aware, but current regressions only cover the failure branch. The success branch should also be locked down so future refactors cannot silently stop forwarding the injected stdin/stdout streams.

## What Changes

- Add a `cmd/lango` regression for broker-mode success with injected stdin/stdout
- Record the seam-forwarding contract in the CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: broker mode forwards configured stdin/stdout seams on success

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
