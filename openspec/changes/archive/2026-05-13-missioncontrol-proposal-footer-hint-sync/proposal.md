## Why

When a proposed mission row is focused, Mission Control's help bar correctly exposes `Enter` accept and `d` dismiss, but the footer hint still falls back to generic request-entry wording. That leaves the footer behind the active row state and makes proposal actions less discoverable than the help bar.

## What Changes

- Make the Mission Control footer hint mention proposal accept/dismiss actions when a proposed mission row is focused.
- Add a regression for the proposal-focused footer hint.
- Sync public docs/spec wording with the footer-level proposal-action guidance.

## Impact

- Improves discoverability for proposal actions without changing behavior.
- Keeps the footer hint aligned with the proposal-focused help bar.
