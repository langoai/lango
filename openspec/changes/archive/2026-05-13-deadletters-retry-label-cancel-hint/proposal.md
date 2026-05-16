## Why

The Dead Letters retry-confirm surface now exposes `r`, `Enter`, and `Esc` in help/docs, but the detail pane's `Retry action:` label still only mentions the confirm keys. That leaves the selected-row status text slightly behind the full runtime contract.

## What Changes

- Expand the retry-confirm label to mention that `Esc` cancels the pending retry request.
- Update regressions that pin the selected-row retry action text.
- Extend the cockpit-pages spec so the confirm-state detail label reflects both confirm and cancel paths.

## Impact

- The selected-row detail pane becomes fully self-describing during retry confirmation.
- Help bar, detail pane, tests, and spec all describe the same confirm/cancel contract.
