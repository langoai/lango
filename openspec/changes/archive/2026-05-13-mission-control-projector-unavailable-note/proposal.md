## Why

`MissionControlPage` already uses degraded notes when the projector returns a degraded snapshot, but when the projector itself is nil the page currently falls back to a plain empty-state snapshot. That looks like "no missions" instead of "mission-control data path is unavailable".

For operator trust, the page should surface the missing dependency explicitly.

## What Changes

- Make `MissionControlPage` emit a degraded note when no projector is configured.
- Add a regression that locks the nil-projector degraded state.
- Sync the cockpit feature docs and cockpit pages spec.

## Impact

- Mission Control no longer disguises an unavailable data path as a normal empty dashboard.
- Degraded-state handling becomes consistent across cockpit pages.
