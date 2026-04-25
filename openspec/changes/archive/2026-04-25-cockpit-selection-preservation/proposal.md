## Why

The cockpit dead-letter page already supports many reload-triggering actions, but it still jumps to the first row after apply, reset, or retry-success refresh even when the operator's current transaction is still present in the refreshed backlog.

## What Changes

- preserve the current selection across apply, reset, and retry-success refresh when the selected transaction remains in the refreshed result set
- keep deterministic fallback behavior:
  - first row when results exist
  - clear selection/detail when the refreshed result set is empty
- document the landed selection-preservation slice in public docs and main OpenSpec specs

## Impact

- smoother operator triage during repeated reloads
- no backend or bridge contract changes
- no change to retry execution semantics
