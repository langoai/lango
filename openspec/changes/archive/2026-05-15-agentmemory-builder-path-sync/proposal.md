## Why

The `agent-memory` main spec still referenced a deleted app-local tool-builder file even though agent memory tools are now owned by the `agentmemory` package and wired from the current app module. That leaves the spec materially stale and sends maintainers to a dead path.

## What Changes

- sync the `agent-memory` main spec to the current builder ownership
- add an executable guard so the deleted app-local builder-path claim cannot silently return

## Impact

- main spec matches the current code layout
- less confusion around agent memory tool registration ownership
- stronger regression protection for deleted builder-path drift
