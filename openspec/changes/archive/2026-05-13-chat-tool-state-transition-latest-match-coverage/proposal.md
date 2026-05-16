## Why

`transitionLatestToolState()` now drives tool rows through approval-related state changes, but there is no regression that proves it updates the most recent matching running row when the same tool name appears more than once in the transcript. That leaves an important selection rule unpinned.

## What Changes

- Add regression coverage for the latest-match selection rule in tool row state transitions.
- Record the selection contract in OpenSpec.

## Impact

- Improves test coverage around a new lifecycle helper.
- Reduces the risk of future regressions when the same tool name appears multiple times in one transcript.
