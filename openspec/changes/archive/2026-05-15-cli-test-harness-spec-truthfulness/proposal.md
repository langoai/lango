## Why

The `cli-test-harness` main spec still described a deleted `internal/testutil/cli_harness.go` file as the shared CLI harness implementation even though the current reusable helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`. That leaves the spec materially stale and makes future contributors chase a path that no longer exists.

## What Changes

- sync the `cli-test-harness` main spec to the current helper layout
- add an executable guard so deleted-harness path claims cannot silently return

## Impact

- main specs better match the actual test helper layout
- less contributor confusion around where shared CLI test utilities live
- stronger regression protection for helper-layout drift
