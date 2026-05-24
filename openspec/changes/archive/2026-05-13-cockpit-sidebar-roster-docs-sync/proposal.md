## Why

The public cockpit feature page and the cockpit sidebar spec still carry stale assumptions from an older roster:

- the sidebar spec still claims the sidebar shows exactly 7 items in the order `Chat, Settings, Tools, Status, Sessions, Tasks, Approvals`
- the feature page does not explain that unavailable optional pages are now shown as disabled destinations rather than as silently missing or fake-live routes

The current cockpit has a 9-item roster headed by Mission Control, and sidebar availability now reflects page registration status directly.

## What Changes

- Update the cockpit feature page to describe current page availability semantics.
- Update the cockpit sidebar spec to require the current 9-item roster and current ordering.
- Add downstream-docs-sync coverage so the public cockpit page stays aligned with the runtime availability model.

## Impact

- Cockpit docs and spec reflect the actual runtime navigation surface.
- Future sidebar work has a current roster contract instead of a stale pre-Mission-Control one.
