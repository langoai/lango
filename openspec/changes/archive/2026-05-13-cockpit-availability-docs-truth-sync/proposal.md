## Why

The current cockpit runtime now registers all first-class sidebar pages, including Dead Letters, and relies on page-level degraded messaging when backing services are absent. But the public cockpit feature page and downstream docs spec still carry wording from the earlier "disabled optional destination" model.

That leaves a direct contradiction between runtime behavior and documentation.

## What Changes

- Update the public cockpit feature page to describe the current always-routable degraded-page model.
- Sync downstream-docs requirements so they no longer require disabled-destination wording for the current runtime surface.

## Impact

- Public cockpit docs match the current runtime navigation contract.
- Downstream docs spec stops enforcing stale wording from an older sidebar-availability model.
