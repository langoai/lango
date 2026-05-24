## Why

Several main specs still referenced single-file paths that no longer exist after refactors or renames. These were not broad architectural notes or globs; they were concrete paths that now send maintainers to dead ends.

## What Changes

- sync the affected main specs to the current paths and package names
- add an executable guard so the known broken single-file references cannot silently return

## Impact

- main specs better match the current repository structure
- fewer dead-end code-path references for maintainers
- stronger regression protection for moved-file drift
