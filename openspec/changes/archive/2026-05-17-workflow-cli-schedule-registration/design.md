## Context

The cron subsystem already persists scheduled jobs through `boot.Storage.Cron()`. Cron job execution delegates to the agent runner, which has access to automation tools including `workflow_run` in runtime wiring. The workflow CLI can therefore register a scheduled workflow without adding a new scheduler subsystem.

## Design

When `lango workflow run <file.flow.yaml> --schedule <cron>` is invoked:

1. Parse and validate the workflow file exactly as today.
2. Build the cron job model and validate its schedule with the same cron parser used by the runtime scheduler.
3. Bootstrap runtime storage.
4. Create a cron job with:
   - name derived from the workflow name, prefixed with `workflow:`
   - schedule type `cron`
   - schedule equal to the effective workflow schedule
   - prompt that explicitly instructs the agent to call `workflow_run` with the absolute workflow file path
   - deliver targets copied from the workflow `deliver_to` list when present
   - enabled state set to true
4. Print a concrete registration confirmation through Cobra command output.

This keeps the CLI thin: it validates input and persists a scheduling request through the existing cron storage facade. Actual execution remains in runtime automation.

## Error Handling

- If bootstrap fails, return `bootstrap: ...` instead of silently pretending registration succeeded.
- If cron storage is unavailable, return a clear workflow schedule registration error.
- If the cron expression cannot be registered by the runtime scheduler, return `invalid workflow schedule: ...` and do not persist a job.
- If job creation fails, wrap it as `register scheduled workflow: ...`.

## Non-Goals

- Do not implement a separate workflow-specific scheduler.
- Do not execute the workflow immediately when `--schedule` is provided.
- Do not invent an API registration endpoint.
