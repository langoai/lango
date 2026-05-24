## Why

The cockpit runtime and feature docs now distinguish degraded page states for Settings, Status, Sessions, Tools, Tasks, Dead Letters, and Approvals. But the README shortcut table and the `lango cockpit` overview in `docs/cli/core.md` still present several of those pages as if they are always fully live.

Those public entrypoints should match the current degraded-page model.

## What Changes

- Update the README cockpit shortcut table for Settings, Status, and Sessions.
- Update the `lango cockpit` overview in `docs/cli/core.md` to explain that core cockpit pages stay routable and surface degraded messaging when dependencies are absent.
- Extend downstream docs requirements so both public entrypoints are covered.

## Impact

- The main public cockpit entrypoints describe the same degraded-page contract as the runtime and feature docs.
- Operators get consistent expectations before entering the cockpit.
