## Why

Mission Control rendering is now hardened, but the `MissionControlProjector` still stores many display-facing text fields in raw form inside `MissionView`, `DecisionView`, `LoopView`, and `CollaborationView`. That leaves replay and downstream-consumer paths dependent on renderer-level sanitization instead of the snapshot model itself being display-safe.

## What Changes

- Sanitize display-facing Mission Control projector text at snapshot construction time.
- Add regression coverage for malformed proposal, decision, loop, and collaboration metadata in the projected snapshot.
- Record the snapshot-level replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns Mission Control projector output with the same plain-text baseline already enforced at render time.
- Prevents raw control text from persisting inside projected cockpit snapshot models.
