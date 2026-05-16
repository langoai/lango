## Why

Several remaining P2P operator commands still used boolean `--json` toggles after adjacent P2P surfaces had already moved to explicit `--output table|json` contracts. That left the family inconsistent, allowed invalid output values to reach bootstrap-dependent work, and left docs/specs partially stale.

## What Changes

- change the remaining P2P operator commands with machine-readable output to use `--output table|json`
- reject unknown output values before bootstrap-dependent work
- sync public docs, main specs, and the focused P2P regression guard to the family-wide contract

## Impact

- consistent P2P CLI UX across inspection and guidance-oriented surfaces
- earlier and more actionable failures for invalid output values
- executable protection against stale `--json` regressions across the whole migrated P2P family
