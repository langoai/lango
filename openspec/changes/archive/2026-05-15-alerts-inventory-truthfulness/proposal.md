## Why

The architecture project-structure page and the README internal tree still omitted the `alerts` CLI family from their inventory summaries, even though `lango alerts list/summary` is implemented and documented elsewhere.

## What Changes

- add `cli/alerts/` to `docs/architecture/project-structure.md`
- add `alerts/` to the README internal tree inventory
- add a regression guard so those inventory summaries keep the current alerts command surface

## Impact

- architecture and README inventory docs better match the shipped alerts CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for alerts inventory truthfulness
