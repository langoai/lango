## Why

The public architecture inventory omits `cli/configcmd/` even though the repository ships a dedicated configuration-management package and already documents its command surface elsewhere.

## What Changes

- add a `cli/configcmd/` row to the architecture project-structure inventory
- add an executable guard that requires the current config management surface in that row
- sync the main docs-only and test-coverage specs

## Impact

- more complete architecture inventory docs
- better discoverability of the shipped config management package
- stronger regression protection against future inventory omissions
