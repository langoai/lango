## Why

The cockpit Tools and Sessions pages currently keep advertising `↑/↓` navigation even when there is nothing to navigate: no tool catalog, no categories, no session list source, or no loaded sessions. That makes the help bar less truthful exactly in the degraded and empty states where the operator most needs reliable guidance.

## What Changes

- Hide Tools page navigation help when there is no catalog or no categories to move through.
- Hide Sessions page navigation help when there are no loaded session rows to move through.
- Add regressions and update the relevant page specs and cockpit feature docs.

## Impact

- Empty and unavailable states stop advertising inert navigation keys.
- Runtime help, tests, docs, and spec describe the same context-sensitive navigation contract.
