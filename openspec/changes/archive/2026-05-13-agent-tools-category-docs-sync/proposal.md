## Why

The public docs and `cli-agent-tools-hooks` spec still describe `lango agent tools` as a tool-to-agent assignment command. The actual implementation loads config and reports tool category availability (`CATEGORY`, `ENABLED`, `DESCRIPTION`) with optional `--category` filtering. That mismatch makes the CLI docs materially misleading.

## What Changes

- Rewrite the `lango agent tools` public docs to describe tool category availability instead of sub-agent partitioning.
- Update examples and flags to include the actual `--category` filter and output shape.
- Sync the `cli-agent-tools-hooks` spec to the implemented contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-agent-tools-hooks`: the documented and specified contract for `lango agent tools` now matches the implemented category-availability behavior.

## Impact

- Affected docs: `docs/cli/agent-memory.md`, `docs/cli/index.md`, `docs/features/agent-format.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/cli-agent-tools-hooks/spec.md`
