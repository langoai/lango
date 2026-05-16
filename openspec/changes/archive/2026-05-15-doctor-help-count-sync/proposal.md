## Why

The `doctor` command help text already describes a much broader diagnostic surface than the older `cli-help-text` main spec claimed. That mismatch leaves the spec materially stale even though the implementation and tests are correct.

## What Changes

- sync the `cli-help-text` main spec to the current `doctor --help` check families and total count
- add a regression test that anchors the current help text contract

## Impact

- the main spec matches the actual production CLI
- future expansions or regressions in `doctor --help` are easier to notice
