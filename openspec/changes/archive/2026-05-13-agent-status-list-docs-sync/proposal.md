## Why

The `lango agent status` and `lango agent list` implementations already expose teammate runtime, performance fields, registry counts, and source-based local/remote list output. But the public CLI docs still showed older example output with outdated agent names, missing registry fields, and the wrong local-agent table columns.

## What Changes

- Rewrite the `lango agent status` examples to show the current field layout, including teammate runtime and registry sections.
- Rewrite the `lango agent list` docs to describe `SOURCE`-based local entries and separate remote A2A rows.
- Sync the CLI docs examples to the current implementation contract without changing runtime behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-agent-inspection`: public CLI examples now match the current `agent status` and `agent list` output shape.

## Impact

- Affected docs: `docs/cli/agent-memory.md`
