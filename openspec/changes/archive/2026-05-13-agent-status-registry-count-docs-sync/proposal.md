## Why

The `lango agent status` docs example was updated to the current field layout, but the registry example counts still showed `Builtin Agents: 0`. The implementation counts the embedded default agent inventory in that field, so the example numbers were still wrong even after the broader output-shape sync.

## What Changes

- Correct the `lango agent status` examples so the registry counts match the current implementation model.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-agent-inspection`: the public `agent status` examples now match the current registry count semantics.

## Impact

- Affected docs: `docs/cli/agent-memory.md`
