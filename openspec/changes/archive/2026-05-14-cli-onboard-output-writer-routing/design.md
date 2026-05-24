## Overview

The onboard wizard itself renders through Bubble Tea, but its preset banner, cancel message, and post-save guidance are emitted after command orchestration in regular Go output helpers. Passing an explicit writer into the orchestration and helper functions is enough to make those outputs deterministic.

## Decisions

### Thread the Cobra writer through the onboard flow

`runOnboard(...)` now accepts an `io.Writer`, and `printNextSteps(...)` renders to that writer instead of process-global stdout.

### Keep the TUI runtime unchanged

The Bubble Tea program still owns terminal rendering during the wizard. Only the non-TUI output surrounding it is moved to the command writer.

## Non-Goals

- No redesign of the onboard wizard itself
- No change to Bubble Tea screen rendering behavior
