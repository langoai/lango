## Why

The P2P workspace tool cluster declares required wrapper inputs, but only `p2p_workspace_create` is directly locked by targeted regressions. The remaining workspace entrypoints can still drift back toward generic downstream failures without an explicit test catching that wrapper breakage.

## What Changes

- Add exact missing-parameter regressions for `p2p_workspace_join`, `p2p_workspace_leave`, `p2p_workspace_status`, `p2p_workspace_post`, and `p2p_workspace_read`.
- Update prompt/public docs to state that these required workspace inputs fail at the wrapper boundary.
- Sync the workspace and production-readiness specs to the same fail-closed contract.

## Impact

- `p2p-workspace`: operator-facing workspace entrypoints become explicitly regression-covered.
- `production-readiness`: workspace tool semantics align with the same actionable missing-parameter standard used across other hardened clusters.
