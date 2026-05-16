## Why

Cockpit approval surfaces are now heavily sanitized at render time, but the shared `PendingApprovalRegistry` still stores the raw `ApprovalRequestMsg` payload. That leaves the cockpit-owned pending approval model dependent on renderer-level sanitization instead of the shared pending snapshot itself being display-safe.

## What Changes

- Sanitize display-facing approval request/viewmodel fields at pending-registry storage time.
- Add regression coverage for malformed pending approval text fields.
- Record the shared pending-approval replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns cockpit-owned pending approval storage with the same plain-text baseline already enforced across Mission Control and chat approval surfaces.
- Prevents raw control text from persisting inside the shared pending approval owner.
