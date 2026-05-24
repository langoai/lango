## Why

The `agent-memory` main spec still referenced a deleted app-local tool-builder file even though the tools are now owned by the `agentmemory` package and wired from the current app module. That makes the spec materially stale and sends maintainers to a dead path.

## What Changes

- sync the `agent-memory` main spec to the current builder ownership
- add an executable guard so the deleted builder-path claim cannot silently return

## Impact

- better alignment between the spec and the current code layout
- less confusion about agent memory tool registration ownership
- stronger regression protection for deleted builder-path drift
