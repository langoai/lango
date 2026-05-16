## Why

The security inspection CLI surfaces still used boolean `--json` toggles while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values.

## What Changes

- change security inspection subcommands to use `--output table|json`
- reject unknown output values before bootstrap-dependent work
- sync public docs and CLI security status spec text to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable output selection
