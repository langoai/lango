## Why

The `cli-test-harness` main spec still described a deleted `internal/testutil/cli_harness.go` path as if it were the current shared harness implementation. That mismatch makes the spec materially false and sends contributors to a file that no longer exists.

## What Changes

- sync the `cli-test-harness` main spec to the current helper layout
- add an executable guard so the stale deleted-path claim cannot silently return

## Impact

- the main spec matches the actual repository structure
- less contributor confusion when looking for shared CLI test helpers
- stronger regression protection for stale helper-path docs
