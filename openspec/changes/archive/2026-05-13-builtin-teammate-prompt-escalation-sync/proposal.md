## Why

The built-in teammate runtime no longer uses `transfer_to_agent("lango-orchestrator")` as its default escalation path, and the main specs already require embedded built-in prompt files to avoid that instruction. But several embedded `prompts/agents/*/IDENTITY.md` files still contained the old transfer-oriented escalation wording.

## What Changes

- Remove transfer-oriented escalation instructions from all embedded built-in teammate `IDENTITY.md` files.
- Align their response rules to the current built-in return-control pattern: emit a short visible escalation summary and end the turn.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `sub-agent-default-prompts`: embedded built-in teammate prompts now match the built-in runtime escalation contract already defined in specs.

## Impact

- Affected prompts: `prompts/agents/{operator,navigator,vault,librarian,automator,planner,chronicler,ontologist}/IDENTITY.md`
