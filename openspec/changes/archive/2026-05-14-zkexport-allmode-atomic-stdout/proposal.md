## Why

`zkexport --all` already removes files from the current run when a later circuit export fails, but it still leaks earlier success progress lines to stdout. That makes a failed run look partially successful and weakens the atomic export story.

## What Changes

- buffer `--all` progress output until the full run succeeds
- keep failure stdout empty while preserving stderr failure reporting
- add regression coverage and sync the ZKP export spec

## Impact

- clearer operator UX on failed multi-circuit exports
- stronger atomicity guarantees for automation wrappers that interpret stdout
