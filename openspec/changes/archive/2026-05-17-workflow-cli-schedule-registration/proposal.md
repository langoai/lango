## Why

`lango workflow run --schedule` currently validates the workflow and then stops with "CLI schedule registration is not implemented yet." That is an operator-facing production gap: the CLI already has a cron store and workflow execution tool, but the workflow command does not connect them.

## What Changes

- Register scheduled workflow runs as cron jobs from `lango workflow run --schedule`.
- Store a deterministic automation prompt that instructs the runtime agent to call `workflow_run` for the selected workflow file.
- Reject invalid cron schedules before persisting an enabled job.
- Keep validation-first behavior and command-output routing intact.
- Update public automation docs and OpenSpec specs to remove the stale "not implemented" guidance.

## Impact

- Affected code: `internal/cli/workflow`, `internal/cron`
- Affected downstream docs: `docs/cli/automation.md`
- Affected specs: `cli-workflow-management`, `downstream-docs-sync`
- Verification: focused workflow CLI tests, OpenSpec validation, full Go build/test
