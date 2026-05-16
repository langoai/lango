## Why

The cockpit Settings page is always available, but when no config-profile store is wired it currently behaves like a normal editor with an implicit no-op save path. That is misleading: operators can believe changes were persisted when they were not.

The page should stay available, but saving must fail closed with explicit messaging.

## What Changes

- Make the nil-store embedded save callback return an actionable error instead of silently succeeding.
- Add an explicit unavailable-persistence note to the Settings page view when the profile store is absent.
- Add regressions and sync the cockpit settings-page contract plus public docs wording.

## Impact

- Operators can still inspect and edit settings in degraded mode, but they are clearly told that saving is unavailable.
- The cockpit Settings page joins the other degraded pages in surfacing missing dependencies explicitly instead of silently no-oping.
