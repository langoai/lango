## Overview

The settings editor itself renders through Bubble Tea, but its cancel message and post-save guidance are regular Go output helpers executed before and after the TUI runtime. Passing an explicit writer into the orchestration and helper functions is enough to make those outputs deterministic.

## Decisions

### Thread the Cobra writer through the settings flow

`runSettings(...)` now accepts an `io.Writer`, and `printNextSteps(...)` renders to that writer instead of process-global stdout.

### Keep the TUI runtime unchanged

The Bubble Tea program continues to own terminal rendering during the editor session. Only the non-TUI output surrounding it is moved to the command writer.

## Non-Goals

- No redesign of the settings editor itself
- No change to Bubble Tea screen rendering behavior
