## Why

The remaining read-only P2P inspection commands still used boolean `--json` toggles even after the first inspection subset moved to explicit `--output table|json` contracts. That left the operator UX inconsistent across adjacent commands and still allowed invalid output values to travel into bootstrap-dependent work.

## What Changes

- change `lango p2p discover`, `pricing`, `reputation`, and `session list` to use `--output table|json`
- reject unknown output values before bootstrap-dependent work for those commands
- sync public docs, main specs, and focused P2P regression guards to the normalized contract

## Impact

- more consistent read-only P2P operator UX
- earlier failures for invalid output values
- broader executable protection against stale `--json` regressions in P2P docs and production code
