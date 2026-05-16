## Why

The top-level `chat` and `cockpit` commands already reject non-interactive startup, but that contract is not currently locked down by dedicated regressions. A future refactor could accidentally weaken the error text or route.

## What Changes

- Add regressions for non-interactive `lango cockpit`
- Add regressions for non-interactive `lango chat`
- Record the contract in CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: chat and cockpit non-interactive startup errors are regression-covered

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
