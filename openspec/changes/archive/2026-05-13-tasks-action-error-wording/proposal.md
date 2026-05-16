## Why

The Tasks page already surfaces cancel/retry outcomes as transient status messages, but failures still show a generic `Error: ...` prefix. That is less specific than the rest of the cockpit's operator-facing failure wording and does not clearly tie the error to a task action.

## What Changes

- Change task action failure status messages from generic `Error:` wording to explicit task-action failure wording.
- Update regressions for cancel/retry failure messaging.
- Extend the task-surface spec and cockpit docs to describe the transient failure message contract.

## Impact

- Operators can immediately tell that the transient error came from a task action.
- Runtime messaging, tests, docs, and spec use the same failure wording.
