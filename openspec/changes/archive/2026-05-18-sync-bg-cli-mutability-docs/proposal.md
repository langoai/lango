## Why

The background automation docs describe the `lango bg` CLI as read-only even though `lango bg cancel <id>` is implemented and mutates a pending or running task by requesting cancellation. That wording conflicts with the command table on the same page and can mislead users about what the CLI actually does.

## What Changes

- update the background automation docs to describe `lango bg` as in-process management rather than read-only management
- explicitly call out that list/status/result inspect state while cancel mutates an eligible task
- add an executable docs-quality guard so the stale read-only wording cannot return
- sync the downstream-docs and test-coverage specs with the corrected contract

## Impact

- public background-task docs match the implemented CLI surface
- the existing root-CLI/server in-memory boundary caveat remains intact
- regression coverage catches future drift in this wording
