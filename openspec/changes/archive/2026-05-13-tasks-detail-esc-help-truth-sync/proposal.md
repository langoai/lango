## Why

The Tasks page detail panel already treats `Esc` as a close action, and the cockpit feature docs describe it that way, but the help bar still labels `Esc` as `back`. That wording is slightly vaguer than the real behavior.

## What Changes

- Change the Tasks page detail-mode `Esc` help label from `back` to `close detail`.
- Add regression coverage for the updated help text.
- Update the task-surface spec to require the close-specific `Esc` label.

## Impact

- The help bar describes the detail-mode `Esc` action precisely.
- Runtime help, tests, docs, and spec align on the same wording.
