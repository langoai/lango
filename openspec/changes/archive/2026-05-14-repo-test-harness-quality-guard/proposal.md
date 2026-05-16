## Why

We already added a CLI-specific guard against global stdio replacement and legacy `testutil.ExecCmd*` helpers, but the underlying hygiene rule is broader than CLI tests. Any repository test that reintroduces those patterns can degrade determinism and pull the suite back toward process-global interception.

## What Changes

- Add an executable repository test that rejects global stdio reassignment across test files under `cmd/` and `internal/`
- Reject legacy `testutil.ExecCmd(...)` and `testutil.ExecCmdOK(...)` references across repository test files
- Record the broader guard in test-infrastructure and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `test-infrastructure`: repository tests are guarded against legacy exec helpers and global stdio reassignment
- `test-coverage`: broader test-harness hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/repo_test_harness_quality_guard_test.go`
- Affected specs: `openspec/specs/test-infrastructure/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
