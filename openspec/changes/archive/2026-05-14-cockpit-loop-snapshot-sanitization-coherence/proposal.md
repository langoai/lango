## Why

Mission Control projector snapshot sanitization now covers missions, decisions, collaboration, and activities, but `LoopView` titles, summaries, and next actions are still copied out raw from the loop projector. That leaves one replayable snapshot path where malformed loop text can persist until render-time sanitization.

## What Changes

- Sanitize projected `LoopView` text fields at snapshot construction time.
- Add regression coverage for malformed loop text in projected snapshots.
- Record the expanded projector replay-safety contract in OpenSpec and downstream docs.

## Impact

- Closes the remaining raw loop-text path inside Mission Control projected snapshots.
- Keeps downstream replay consumers aligned with the same plain-text baseline as rendered Mission Control lanes.
