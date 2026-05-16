## Why

The Approvals page always advertises `↑/↓` navigation in the help bar even when the current section has nothing to move through: no history entries, no grants, or an unavailable store. That makes the help bar less truthful in exactly the degraded and empty states where the operator needs reliable guidance.

## What Changes

- Hide Approvals section navigation help when the active section has no navigable rows.
- Keep the section-toggle and revoke bindings only when they are meaningful.
- Add regressions and update the approval-history-view spec plus cockpit docs.

## Impact

- Empty and unavailable Approvals states stop advertising inert navigation keys.
- Runtime help, tests, docs, and spec describe the same context-sensitive contract.
