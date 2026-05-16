## Why

The approval CLI still used a boolean `--json` toggle while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX less predictable and prevented fail-fast validation of invalid output values.

## What Changes

- change `lango approval status` to use `--output table|json`
- reject unknown output values before config loading
- add public CLI documentation for the normalized approval output contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable output selection
