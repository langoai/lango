## Why

The dead-letter operator surfaces can already request retries from both CLI and cockpit, but the current retry feedback is too coarse. Operators need clearer wording for retry acceptance, precheck rejection, and request failure so they can understand what happened immediately after invoking recovery.

## What Changes

- refine CLI retry success/failure wording
- refine CLI structured retry result output
- refine cockpit retry state-transition wording
- refine cockpit retry success/failure wording
- sync public docs and main OpenSpec requirements for the landed recovery UX

## Impact

- clearer operator recovery semantics without changing the retry control plane
- better parity between cockpit and CLI retry behavior
- more truthful public docs for recovery outcomes
