## Why

The librarian CLI family still used a boolean `--json` toggle while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values.

## What Changes

- change `lango librarian status` and `lango librarian inquiries` to use `--output table|json`
- reject unknown output values before config or bootstrap loading
- add public librarian CLI docs for the normalized contract

## Impact

- more consistent operator UX across CLI surfaces
- clearer machine-readable output selection
- earlier failures for invalid output values
