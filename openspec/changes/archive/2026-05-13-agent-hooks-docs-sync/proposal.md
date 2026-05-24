## Why

The `lango agent hooks` implementation already renders a configuration summary plus a registry snapshot with pre/post hooks, wirable state, and KnowledgeSaveHook details. But the public CLI docs still showed an older flat `HOOK / TYPE / STATUS` example that no longer matches the command output.

## What Changes

- Update the `lango agent hooks` docs example to match the current text output shape.
- Mention the JSON surface (`registry.preHooks`, `registry.postHooks`, `details`) so operators know what the structured output contains.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-agent-tools-hooks`: public docs now match the current `agent hooks` output contract.

## Impact

- Affected docs: `docs/cli/agent-memory.md`
