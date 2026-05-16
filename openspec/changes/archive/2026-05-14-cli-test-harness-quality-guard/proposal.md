## Why

We already removed process-global stdio interception and legacy `testutil.ExecCmd*` usage from CLI regressions, but nothing prevents those brittle patterns from being reintroduced later. That would quietly undo the command-stream testing cleanup.

## What Changes

- Add an executable repository test that rejects process-global stdio replacement in CLI tests
- Reject legacy `testutil.ExecCmd(...)` and `testutil.ExecCmdOK(...)` references in CLI tests
- Record the guard in CLI test-harness and test-coverage specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-test-harness`: CLI regressions are guarded against global-stdio and legacy ExecCmd fallback
- `test-coverage`: CLI harness hygiene has an executable guard

## Impact

- Affected code: `internal/testutil/cli_test_harness_quality_guard_test.go`
- Affected specs: `openspec/specs/cli-test-harness/spec.md`, `openspec/specs/test-coverage/spec.md`
- No runtime behavior changes
