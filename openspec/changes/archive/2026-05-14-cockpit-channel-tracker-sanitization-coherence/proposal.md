## Why

The cockpit context panel now sanitizes channel names at render time, but `ChannelTracker` still stores the raw channel label inside its `channelStatus` snapshot. That leaves a replayable intermediate model dependent on renderer-level sanitization instead of the snapshot itself being display-safe.

## What Changes

- Sanitize channel names when seeding and recording channel tracker snapshots.
- Add regression coverage for malformed channel labels.
- Record the channel snapshot replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns `channelStatus` storage with the same plain-text baseline already enforced in the context panel.
- Prevents raw control text from persisting inside cockpit channel snapshots.
