## Why

The architecture project-structure page and the README internal tree still described only a subset of the current P2P CLI surface. Workspace, git, provenance, and some higher-level team/ZKP surfaces were already shipped but not reflected in those inventory summaries.

## What Changes

- update `docs/architecture/project-structure.md` so `cli/p2p/` includes workspace, git, provenance, team, and ZKP surfaces
- update the README internal tree inventory to include `workspace`, `git`, and `provenance`
- add a regression guard so those inventory summaries keep the current P2P command surface

## Impact

- architecture and README inventory docs better match the shipped P2P CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for P2P inventory truthfulness
