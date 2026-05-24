## Why

`lango workflow run` now routes non-error output through the command writer, but its direct-success path still lacked command-level regression coverage. That leaves the completion banner and per-step output weakly verified compared with the other workflow branches.

## What Changes

- Add a small direct-execution seam for deterministic workflow-run success tests
- Add command-level regression coverage for successful direct execution output
- Update docs and OpenSpec to make the success-path output contract explicit

## Impact

- Improves confidence in the workflow-run happy path without changing workflow runtime behavior
- Completes coverage of the recently hardened workflow-run command stream
