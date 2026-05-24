## Why

The Tasks page help bar is already context-sensitive for scroll and action keys, but `Enter` still shows `details` even when the detail panel is already open. In that state the real action is to close the panel, so the help wording lags the actual behavior.

## What Changes

- Make the Tasks page `Enter` help label reflect `close detail` while detail mode is open.
- Add regressions for the context-sensitive `Enter` help wording.
- Update the task-surface spec to require the detail-mode help label to match the close action.

## Impact

- The help bar describes the current `Enter` action more accurately.
- Runtime help, tests, and spec all align on the detail toggle contract.
