## Why

The cockpit Tasks page currently uses the same empty-state wording for two different situations:

- the background task manager is not configured at all
- the task manager exists but currently has no tasks

Those are operationally different states, and the page should not collapse them into one message.

## What Changes

- Show explicit unavailable messaging when `TasksPage` has no background-task lister.
- Preserve the existing empty-state wording only for the configured-but-empty case.
- Add regression coverage and sync public docs plus the task-surface spec.

## Impact

- Operators can distinguish missing automation wiring from an idle task queue.
- Cockpit degraded-page messaging stays consistent across Tasks, Approvals, and Dead Letters.
