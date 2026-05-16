## Why

The P2P inspection commands still used boolean `--json` toggles even though newer operator-facing CLI surfaces now expose explicit `--output table|json` contracts. That inconsistency made machine-readable output selection harder to predict and prevented fail-fast validation of invalid output values before bootstrap work.

## What Changes

- change `lango p2p status`, `peers`, and `identity` to use `--output table|json`
- reject unknown output values before bootstrap loading for that inspection subset
- sync public docs, main specs, and focused regression guards to the normalized contract

## Impact

- more consistent operator UX across inspection-oriented CLI surfaces
- earlier and more actionable failures for invalid output values
- executable protection against stale docs or boolean `--json` regressions in the migrated subset
