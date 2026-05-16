## Why

In the standalone workbench surface, Mission Control's empty-state `Enter` key is not always a generic submit action. When the composer is empty it seeds the default starter prompt, and once a starter is armed it runs that starter. The current help label still says `submit` in both cases.

## What Changes

- Make the empty-workbench `Enter` help label context-sensitive: seed when the default starter will be staged, run when a starter is already armed.
- Add regressions for both empty-workbench states.
- Update cockpit page specs and feature docs to describe the current workbench-empty `Enter` contract.

## Impact

- The help bar matches the actual empty-workbench behavior instead of collapsing distinct actions into `submit`.
- Runtime help, tests, docs, and spec all describe the same starter flow.
