## Why

The `production-readiness` spec still contains two overlapping requirement blocks for RPC wallet address cleanup. That duplication adds noise without adding new behavioral coverage.

## What Changes

- Remove the duplicate RPC wallet address-cleanup requirement block.
- Keep the surviving requirement fully covering timeout and companion-error cleanup.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: RPC wallet address cleanup is stated once, with no duplicate requirement block.

## Impact

- Affected specs: `openspec/specs/production-readiness/spec.md`
