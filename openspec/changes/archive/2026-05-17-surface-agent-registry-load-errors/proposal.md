# Surface Agent Registry Load Errors

## Why

`lango agent list` and `lango agent status` inspect embedded and user-defined agent registry state. User-defined agent store errors are currently discarded, so an invalid `AGENT.md` can make user agents disappear from CLI output while the command still exits successfully.

## What Changes

- Return actionable errors when configured user-defined agent registry files cannot be loaded.
- Return embedded registry load errors from agent status instead of reporting misleading counts.
- Preserve existing behavior for missing optional `agent.agentsDir`.
- Update public agent CLI documentation to describe invalid user agent file handling.
