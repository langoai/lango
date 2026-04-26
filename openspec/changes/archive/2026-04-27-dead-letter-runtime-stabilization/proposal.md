## Why

Several high-severity stabilization gaps remained after the larger dead-letter and runtime workstreams:

- cockpit dead-letter filters were truncated at the shell adapter
- CLI and cockpit retry could invoke replay with no principal
- background task execution had no panic recovery
- reputation updates could lose concurrent writes
- `clampScore` could pass `NaN`
- cockpit dispatch-family grouping could drift from the CLI classifier

## What Changes

- Forward the full dead-letter filter set from `cmd/lango` into the cockpit dead-letter bridge.
- Inject a default local principal into CLI and cockpit retry invocations when the runtime context is otherwise empty.
- Recover from panics in the background manager and fail tasks explicitly instead of orphaning them.
- Serialize reputation updates per peer.
- Clamp `NaN` scores to `0`.
- Reuse the shared dispatch-family classifier in cockpit summary aggregation.
- Truth-align the dead-letter docs and docs-only OpenSpec requirements for the stabilized behavior.

## Impact

- Cockpit filters and summaries are now computed from the intended backlog query.
- Retry surfaces no longer fail immediately on empty actor context.
- Background task failures are more survivable under panic conditions.
- Reputation persistence is less vulnerable to lost updates.
