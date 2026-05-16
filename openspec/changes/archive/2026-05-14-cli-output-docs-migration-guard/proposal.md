## Why

Several operator-facing CLI families have already migrated from `--json` toggles to explicit `--output table|json` contracts. Without a docs guard, public pages and main specs can silently drift back to the old examples even while the code remains correct.

## What Changes

- add an executable repository guard that rejects stale `--json` docs for migrated CLI families
- record that regression boundary in docs-only and test-coverage specs

## Impact

- cheaper detection of public-doc and spec drift
- stronger alignment between implemented CLI UX and operator-facing documentation
