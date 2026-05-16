## Why

The Approvals page already distinguishes the fully unavailable state (`history == nil && grants == nil`) from the empty configured state. But partial degraded states are still misleading:

- missing history store shows `No history entries`
- missing grant store shows `No active grants`

Those messages imply configured-but-empty data rather than absent wiring.

## What Changes

- Show explicit section-level unavailable messaging when only one approval store is absent.
- Add regressions for history-only and grants-only degraded states.
- Sync the cockpit docs and approvals spec accordingly.

## Impact

- The Approvals page now distinguishes full unavailable, partial unavailable, and empty configured states cleanly.
- Operators get accurate section-level wiring feedback instead of misleading empty-state copy.
