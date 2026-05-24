## Why

The Approvals page footer already describes `R` as revoking the selected session, and the runtime path calls `GrantStore.RevokeSession(...)`. The help bar still labels `R` as `revoke all`, which is less precise and inconsistent with the footer.

## What Changes

- Rename the `R` help-bar label to `revoke session`.
- Add a regression covering the label.
- Sync public docs/spec wording with the unified label.

## Impact

- Removes a visible wording mismatch from the Approvals page.
- Makes the `R` action describe its real session-scoped effect more clearly.
