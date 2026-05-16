## Why

The P2P git tool cluster declares required wrapper inputs, but there is no direct regression coverage that locks the exact missing-parameter behavior for those entrypoints. That leaves room for drift back toward generic downstream git-service failures before repository lookup or diff resolution begins.

## What Changes

- Add exact missing-parameter regressions for `p2p_git_init`, `p2p_git_push`, `p2p_git_log`, `p2p_git_diff`, and `p2p_git_leaves`.
- Update prompt/public docs to state that these required git inputs fail at the wrapper boundary.
- Sync the workspace and production-readiness specs to the same fail-closed contract.

## Impact

- `p2p-workspace`: git bundle entrypoints become explicitly regression-covered.
- `production-readiness`: workspace git tool semantics align with the same actionable missing-parameter standard used elsewhere.
