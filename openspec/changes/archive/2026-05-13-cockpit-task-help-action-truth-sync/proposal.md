## Why

The cockpit Tasks page help bar currently advertises `c` and `r` whenever a `TaskActioner` exists, even if the selected task state does not support cancel or retry. The key handlers already no-op on invalid states, but the help still overpromises.

The help bar should reflect the currently selected task's actual action surface.

## What Changes

- Make Tasks page help show `c` only for running/pending tasks.
- Make Tasks page help show `r` only for failed/cancelled tasks.
- Leave action keys hidden for states where no action is available.
- Add regressions and sync docs/spec wording.

## Impact

- The Tasks page help bar becomes state-accurate instead of misleading.
- Operators see only the action keys that actually do something for the selected row.
