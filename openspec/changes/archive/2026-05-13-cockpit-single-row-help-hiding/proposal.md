## Why

Tools and Sessions already hide navigation help when there is nothing at all to move through, but they still advertise `↑/↓` when there is exactly one category or one session row. In that state the keys are still inert.

## What Changes

- Show Tools navigation help only when there are at least two categories.
- Show Sessions navigation help only when there are at least two loaded session rows.
- Add regressions and update the relevant specs and cockpit docs.

## Impact

- The help bar only advertises vertical navigation when there is an actual alternative row to move to.
- Runtime help, tests, docs, and spec use the same stricter contract.
