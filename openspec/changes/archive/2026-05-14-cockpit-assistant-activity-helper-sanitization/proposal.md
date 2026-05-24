## Why

The cockpit activity buffer now sanitizes summaries at append time, but the exported helper `NewAssistantSummaryActivity()` still returns a raw `MissionActivityItem` summary before buffering. The workbench reuses that helper directly, so this exported boundary can still hand raw control text to sibling consumers.

## What Changes

- Normalize assistant activity summaries before `NewAssistantSummaryActivity()` returns them.
- Add regression coverage for malformed helper output.
- Record the exported helper replay-safety contract in OpenSpec and downstream docs.

## Impact

- Makes the shared assistant-activity helper safe for direct reuse outside the activity buffer append path.
- Removes the remaining raw summary escape hatch in the Mission Control/workbench shared activity helper.
