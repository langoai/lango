## Why

The architecture project-structure page and the README internal tree still omitted several shipped CLI families even after earlier inventory sync work. Chat, extension, provenance, run, sandbox, and status were implemented and documented elsewhere but missing from those inventory summaries.

## What Changes

- add chat, extension, provenance, run, sandbox, and status rows to `docs/architecture/project-structure.md`
- add the same surfaces to the README internal tree inventory
- add a regression guard so those inventory summaries keep the current command surface

## Impact

- architecture and README inventory docs better match the shipped CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for inventory truthfulness
