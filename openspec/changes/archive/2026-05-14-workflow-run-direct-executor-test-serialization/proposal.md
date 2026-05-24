## Why

`internal/cli/workflow/workflow_run_schedule_test.go` mutates a package-global direct-execution seam while also running in parallel. That makes `go test ./...` nondeterministic and can produce false failures unrelated to the actual change under test.

## What Changes

- Serialize workflow run schedule tests that rely on the package-global direct-execution seam
- Keep behavioral assertions unchanged; only remove the race/flakiness source
- Record the test-stability contract in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `test-coverage`: workflow run command regressions remain deterministic when package-global seams are involved

## Impact

- Affected code: `internal/cli/workflow/workflow_run_schedule_test.go`
- Affected specs: `openspec/specs/test-coverage/spec.md`
- No runtime behavior change; this is test-suite stabilization
