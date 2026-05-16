## Why

`internal/testutil/cli_harness.go` provided a global stdout interception helper for Cobra tests, but all known internal call sites have been migrated to package-local command writer helpers. The shared helper is now dead code and preserves an unsafe pattern we no longer want to keep around.

## What Changes

- Remove the unused `internal/testutil/cli_harness.go` helper
- Record the cleanup in OpenSpec test-coverage artifacts

## Impact

- Simplifies the shared test utility surface
- Removes a global stdout interception pattern that is no longer needed
