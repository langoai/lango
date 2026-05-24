## Why

`lango workflow run --schedule` currently tells operators to POST to `/api/workflow/register`, but no such public server surface exists. That is a real UX contract bug: the CLI points users to a nonexistent path instead of telling them the registration flow is not implemented yet.

## What Changes

- Change `workflow run --schedule` output to say CLI schedule registration is not implemented yet.
- Direct operators toward `lango cron add` or runtime automation tools instead of a nonexistent endpoint.
- Add a CLI regression proving the scheduled path no longer mentions `/api/workflow/register`.
- Sync the CLI automation docs and workflow management spec to the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-workflow-management`: scheduled `workflow run` now fails closed with truthful guidance instead of pointing to a nonexistent endpoint.

## Impact

- Affected code: `internal/cli/workflow/workflow.go`
- Affected tests: `internal/cli/workflow/workflow_run_schedule_test.go`
- Affected docs: `docs/cli/automation.md`
- Affected specs: `openspec/specs/cli-workflow-management/spec.md`
