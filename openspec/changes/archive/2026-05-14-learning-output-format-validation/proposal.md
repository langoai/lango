## Why

The learning CLI family still used a boolean `--json` toggle while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values.

## What Changes

- change `lango learning status` and `lango learning history` to use `--output table|json`
- reject unknown output values before config or bootstrap loading
- sync public docs and CLI learning inspection spec to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- cleaner wrapper expectations around machine-readable output
