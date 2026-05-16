## Why

Mission Control now exposes `d` in the help bar for selected proposed missions, but the actual handler only honors `d` while the missions lane is focused. When focus is on decisions or the composer, the help can still advertise a dismiss action that won't fire.

## What Changes

- Restrict the `d` help binding to the missions-focused proposed-row state.
- Add regressions for mission-focus vs non-mission-focus help behavior.
- Update the cockpit-pages spec and feature docs to reflect the focus-sensitive dismiss contract.

## Impact

- The help bar stops advertising a proposal-dismiss key when it cannot actually run.
- Runtime help, tests, docs, and spec describe the same focus-sensitive behavior.
