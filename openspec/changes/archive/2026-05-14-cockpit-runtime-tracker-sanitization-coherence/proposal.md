## Why

The cockpit context panel now sanitizes runtime active-agent labels at render time, but `RuntimeTracker` still stores the raw active-agent string in its `runtimeStatus` snapshot. That leaves a replayable intermediate model dependent on renderer-level sanitization instead of the snapshot itself being display-safe.

## What Changes

- Sanitize active-agent labels when recording runtime tracker snapshots.
- Add regression coverage for malformed delegation and active-agent labels.
- Record the runtime snapshot replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns `runtimeStatus` storage with the same plain-text baseline already enforced in the context panel.
- Prevents raw control text from persisting inside cockpit runtime snapshots.
