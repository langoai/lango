## Why

The graph CLI family still used a boolean `--json` toggle while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values.

## What Changes

- change graph CLI inspection and mutation commands to use `--output table|json`
- reject unknown output values before config loading or file parsing
- sync public docs and graph CLI spec text to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable output selection
