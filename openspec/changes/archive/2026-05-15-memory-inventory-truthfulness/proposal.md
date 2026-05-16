## Why

The architecture project-structure page and the README internal tree still described only a subset of the current memory CLI surface. The shipped `lango memory agents` and `lango memory agent <name>` commands were already documented elsewhere but missing from those inventory summaries.

## What Changes

- update `docs/architecture/project-structure.md` so `cli/memory/` includes `agents` and `agent <name>`
- update the README internal tree inventory to include the same memory surface
- add a regression guard so those inventory summaries keep the current memory command surface

## Impact

- architecture and README inventory docs better match the shipped memory CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for memory inventory truthfulness
