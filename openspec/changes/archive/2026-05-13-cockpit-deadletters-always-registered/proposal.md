## Why

The cockpit already treats Tools, Status, Settings, Tasks, and Approvals as routable degraded surfaces when their optional backing dependencies are absent. `DeadLettersPage` also already has explicit nil-function failure messaging, but startup wiring still hides the page behind bridge readiness.

That makes one cockpit destination behave differently from the rest of the degraded UX model even though the page can explain its unavailable state directly.

## What Changes

- Always register the Dead Letters cockpit page.
- Keep bridge-backed callbacks when available, but fall back to a degraded page with explicit unavailable messaging when the bridge is absent.
- Add regressions for always-registered wiring and nil-list activation behavior.
- Sync docs/spec wording so Dead Letters is described like the other degraded cockpit pages.

## Impact

- Cockpit navigation becomes more consistent: operators can always open Dead Letters.
- Unavailable dead-letter tooling is surfaced as an explicit page state instead of a missing route.
