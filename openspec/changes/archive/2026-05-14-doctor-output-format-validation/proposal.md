## Why

The doctor CLI still used a boolean `--json` toggle while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX less predictable and prevented fail-fast validation of invalid output values before bootstrap.

## What Changes

- change `lango doctor` to use `--output table|json`
- reject unknown output values before bootstrap
- sync README, CLI docs, and doctor spec text to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable diagnostic output selection
