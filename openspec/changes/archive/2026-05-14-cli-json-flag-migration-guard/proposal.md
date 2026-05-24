## Why

Several operator-facing CLI families have now been migrated from boolean `--json` toggles to explicit `--output table|json` contracts. Without a guard, future changes could silently reintroduce the older UX in those same families and undo the consistency work.

## What Changes

- add an executable repository guard that rejects boolean `--json` flag declarations inside migrated CLI families
- record that regression boundary in the main test-coverage spec

## Impact

- cheaper detection of CLI UX regressions
- stronger consistency across migrated operator-facing command families
