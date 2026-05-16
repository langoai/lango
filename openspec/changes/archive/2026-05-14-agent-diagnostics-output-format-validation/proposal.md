## Why

The agent diagnostics subset still used boolean `--json` toggles while newer operator-facing commands have moved to explicit `--output table|json` contracts. That inconsistency made the UX harder to predict and prevented fail-fast validation of invalid output values before bootstrap.

## What Changes

- change `lango agent trace list`, `trace show`, `graph`, and `trace metrics` to use `--output table|json`
- reject unknown output values before bootstrap-dependent work
- sync docs and agent diagnostics spec text to the normalized output-format contract

## Impact

- more consistent operator UX across CLI surfaces
- earlier failures for invalid output values
- clearer machine-readable diagnostic output selection
