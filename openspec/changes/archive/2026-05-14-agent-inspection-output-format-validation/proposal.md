## Why

The agent inspection CLI subset still used boolean `--json` toggles while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values.

## What Changes

- change `lango agent status`, `list`, `tools`, and `hooks` to use `--output table|json`
- reject unknown output values before config loading
- sync public docs and agent inspection spec text to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable inspection output selection
