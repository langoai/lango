## Why

Cockpit sidebar entries are currently rendered from a static page list even when some pages are not actually registered. In those cases the UI shows navigable destinations that silently refuse to open because `switchPage()` rejects missing pages.

That is a degraded UX contract: unavailable cockpit pages should be visibly disabled, not look live and then no-op.

## What Changes

- Add sidebar availability syncing so unregistered optional pages render disabled.
- Re-enable sidebar items automatically when their pages are registered.
- Add regression coverage for disabled-by-default optional pages and enable-on-register behavior.

## Impact

- Cockpit navigation better matches actual runtime wiring.
- Operators stop seeing live-looking sidebar destinations that cannot open.
