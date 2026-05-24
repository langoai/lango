## Why

The README cockpit shortcut table still describes `Tasks` and `Approvals` as if they are always fully live surfaces. The runtime and cockpit feature docs now distinguish:

- `Tasks` degrades to unavailable messaging when the background task manager is absent
- `Approvals` degrades to unavailable messaging when the approval stores are absent

The shortcut table should not lag behind the real operator contract.

## What Changes

- Update the README cockpit shortcut table for `Tasks` and `Approvals`.
- Extend downstream docs requirements so the README is covered by the current degraded-surface contract.

## Impact

- README matches the current cockpit runtime and feature docs.
- Public operator-facing wording stays consistent across the main entrypoints.
