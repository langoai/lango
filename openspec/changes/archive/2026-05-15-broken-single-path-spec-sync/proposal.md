## Why

Several main specs still referenced single-file paths that no longer exist after package moves and refactors. Those stale references make the specs less trustworthy and slow down maintainers who follow them into dead ends.

## What Changes

- sync the affected specs to the current paths and package names
- add an executable guard so the known-broken single-file references cannot silently return

## Impact

- better alignment between main specs and the current repository layout
- fewer dead-end code-path references for maintainers
- stronger regression protection for moved-file drift
