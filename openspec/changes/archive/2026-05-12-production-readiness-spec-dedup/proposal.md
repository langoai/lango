## Why

The `production-readiness` spec accumulated overlapping requirements while we tightened RPC wallet cleanup coverage. That duplication makes the spec noisier than necessary and weakens it as a trustworthy contract.

## What Changes

- Consolidate overlapping RPC wallet address-cleanup requirements into one requirement block.
- Preserve every actual behavior contract while removing repeated phrasing.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: The spec now states the RPC wallet address cleanup contract once, without duplicated requirement blocks.

## Impact

- Affected specs: `openspec/specs/production-readiness/spec.md`
