## Why

The Approvals page already hides inert navigation help in empty sections, but it still advertises `↑/↓` when the active section has exactly one row. In that state there is nothing to move to, so the help is still overstating what the operator can do.

## What Changes

- Show Approvals navigation help only when the active section has at least two rows.
- Keep section-toggle and revoke actions available when they still make sense.
- Add regressions and update the approval-history-view spec plus cockpit docs.

## Impact

- Single-row Approvals states stop advertising inert navigation keys.
- Runtime help, tests, docs, and spec describe the same stricter contract.
