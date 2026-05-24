## Why

The public cockpit docs and the main `cockpit-sessions-page` spec already describe the Sessions page as a newest-first summary list, but the runtime currently renders sessions in whatever order the list function returns. That leaves ordering nondeterministic and can surface stale sessions above more recent ones.

## What Changes

- Sort loaded session summaries by `UpdatedAt` descending inside the Sessions page before rendering.
- Add regression coverage for unsorted inputs.
- Keep the main sessions-page spec and public cockpit docs aligned with the enforced runtime contract.

## Impact

- Sessions page ordering becomes deterministic and matches the documented newest-first contract.
- Operators see the most recent session activity first without relying on upstream callers to pre-sort.
