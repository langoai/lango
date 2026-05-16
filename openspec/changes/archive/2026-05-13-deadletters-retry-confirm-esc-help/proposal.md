## Why

The Dead Letters page already lets the operator cancel a pending retry confirmation with `Esc`, but that action is currently undiscoverable from the help bar. A visible two-step retry flow should also expose its cancel path.

## What Changes

- Show an `Esc` help binding while a Dead Letters retry confirmation is pending.
- Add regressions for the confirm-state help contract.
- Update cockpit specs and feature docs to describe the confirm-cancel key.

## Impact

- Operators can discover how to back out of a pending retry request without guessing.
- Runtime help, tests, docs, and spec all describe the same retry-confirm workflow.
