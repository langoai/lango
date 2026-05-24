## Why

The main `agent-routing` spec still says the builtin `agentSpecs` registry contains 6 sub-agents, but the actual runtime has 8 built-in teammate types: `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. That makes the authoritative routing spec stale against the current runtime and public docs.

## What Changes

- Update the `agent-routing` main spec so the builtin registry count and ordered teammate list match the current runtime.

## Impact

- `agent-routing`: the canonical spec matches the actual `agentSpecs` registry again.
