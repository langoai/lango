## Why

The Tasks page detail panel already uses `↑/k` and `↓/j` to scroll detail content, and public cockpit docs say so, but the help bar still labels those bindings as generic `up` and `down`. That leaves the in-product guidance less precise than the actual behavior.

## What Changes

- Make Tasks page help label `↑/k` and `↓/j` as scroll controls while detail mode is open.
- Add regressions for the detail-mode help text.
- Update the task-surface spec to require the scroll-specific help wording.

## Impact

- The help bar explains detail-panel behavior accurately.
- Runtime help, tests, docs, and spec all describe the same detail-mode key contract.
