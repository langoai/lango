## Why

Mission Control already treats `Enter` on a focused proposed mission row as an accept action, but the help bar still labels `Enter` generically as `submit`. That understates the actual action on a high-visibility proposal path.

## What Changes

- Render `Enter` as `accept` when a proposed mission is selected and the missions lane is focused.
- Add regressions for proposal-row vs non-proposal help labeling.
- Update the cockpit-pages spec and cockpit feature docs to describe the same context-sensitive `Enter` behavior.

## Impact

- Proposal acceptance becomes discoverable directly from the help bar.
- Runtime help, tests, docs, and spec align on the current proposal-row `Enter` contract.
