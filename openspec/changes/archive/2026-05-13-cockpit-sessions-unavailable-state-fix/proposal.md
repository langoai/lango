## Why

The cockpit Sessions page currently conflates two different states:

- no session-list source is configured
- the session list source exists but there are no sessions

The page should distinguish unavailable wiring from an empty session history, just like the other degraded cockpit pages now do.

## What Changes

- Render explicit unavailable messaging when `SessionsPage` has no list function.
- Preserve `No sessions found.` only for the configured-but-empty case.
- Add regression coverage and sync cockpit docs plus the sessions-page spec.

## Impact

- Operators can distinguish missing session-list wiring from an empty session history.
- Sessions page behavior becomes consistent with Tasks, Approvals, and Dead Letters.
