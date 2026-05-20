## Why

Two cockpit main specs still lag behind the actual page behavior:

- `tui-task-surface` still says nil task-manager renders `No background tasks available`
- `approval-history-view` still has an ambiguous "no history and no grants exist" scenario that does not reflect the now-separated unavailable vs configured-empty states

These are direct code/spec mismatches, not just docs drift.

## What Changes

- Update `tui-task-surface` to require the current nil-manager unavailable message.
- Tighten `approval-history-view` so its generic empty-state scenario explicitly refers to configured stores with no data.

## Impact

- Main specs align with current cockpit behavior.
- Future validation no longer rests on stale page copy contracts.
