## Why

Mission Control header rendering is hardened, but `MissionControlProjector` still stores header summaries such as active-agent, provider/model, metrics, and degraded-note text in raw form inside `HeaderView`. That leaves replay and downstream-consumer paths dependent on renderer-level sanitization instead of the snapshot model itself being display-safe.

## What Changes

- Sanitize Mission Control header summary text at snapshot construction time.
- Add regression coverage for malformed projected header summaries.
- Record the header snapshot replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns `HeaderView` storage with the same plain-text baseline already enforced at render time.
- Prevents raw control text from persisting inside projected Mission Control header snapshots.
