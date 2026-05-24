## Why

The cockpit Tools page already degrades safely when the tool catalog is absent, but its left and right panes use different unavailable messages (`No tool catalog available.` vs `Tool catalog is not available.`). That inconsistency makes the degraded state feel less intentional than the rest of cockpit.

## What Changes

- Use one shared unavailable wording for both sides of the Tools page when the catalog is nil.
- Update regressions and the Tools page spec/docs to reflect the unified message.

## Impact

- The Tools page degraded state reads as one coherent operator surface.
- Runtime, tests, docs, and spec all point at the same unavailable wording.
