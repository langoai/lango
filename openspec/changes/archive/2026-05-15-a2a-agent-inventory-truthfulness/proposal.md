## Why

The architecture inventory and README internal tree still lagged behind the current A2A and agent command surface. README omitted the `a2a` row entirely, abbreviated `agent/` to the inspection subset, and still carried a stale duplicate `chat` row.

## What Changes

- add `a2a` to the README internal tree inventory
- expand the `agent` inventory in README and architecture docs to include diagnostics
- remove the stale duplicate `chat` row from the README internal tree
- add an executable guard so those inventory summaries keep the current A2A/agent surface

## Impact

- architecture and README inventory docs better match the shipped A2A and agent CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for A2A/agent inventory truthfulness
