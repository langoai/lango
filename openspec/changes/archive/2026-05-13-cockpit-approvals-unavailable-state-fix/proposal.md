## Why

The cockpit Approvals page currently shows `No approval history yet.` both when the approval subsystem is genuinely unavailable and when the approval stores exist but are simply empty.

That hides a real operator distinction: "not configured" and "configured but empty" are not the same state.

## What Changes

- Make the Approvals page show explicit unavailable messaging when both approval stores are absent.
- Keep the existing empty-state message only for the configured-but-empty case.
- Add regression coverage and sync public cockpit docs plus the approvals page spec.

## Impact

- Operators can tell whether the approval subsystem is unavailable or just idle.
- Cockpit degraded-page messaging becomes more consistent with the other pages.
