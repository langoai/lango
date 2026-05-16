## Why

The cockpit feature reference explains Approvals and Tasks in detail, but it still lacks a dedicated Sessions section even though the runtime now has a clear contract:

- sessions are listed newest-first
- the page distinguishes unavailable session-list wiring from an empty session history

That behavior should be visible in the public cockpit reference.

## What Changes

- Add a Sessions page section to `docs/features/cockpit.md`.
- Extend downstream docs requirements so the cockpit feature page is expected to describe the Sessions page behavior and unavailable-state split.

## Impact

- The cockpit feature reference covers the Sessions page at the same level as other detail pages.
- Public docs better reflect the current runtime contract for session inspection.
