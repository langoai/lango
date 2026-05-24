## Why

The repository already contains landed quality hardening for deterministic clock-sensitive regressions and for positive duration recording in the parallel read-only executor, but there is no corresponding OpenSpec change record for that baseline. That leaves the quality contract under-documented and makes later audit trails weaker than the code and tests themselves.

## What Changes

- Record the deterministic clock-sensitive regression baseline for proposal TTL and Mission Control agenda freshness tests.
- Record the positive-duration contract for completed parallel read-only executor invocations, including failing handlers.
- Close the leftover archive task ledgers for two already-archived changes whose implementation and validation are complete.

## Impact

- Improves traceability for already-landed quality work without changing runtime behavior.
- Keeps the OpenSpec archive aligned with the actual repository state.
