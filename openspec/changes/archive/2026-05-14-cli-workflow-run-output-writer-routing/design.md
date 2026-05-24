## Overview

`lango workflow run` has one read-only validation/schedule path and one direct execution path. Both already render plain text output inline, so routing them through the command writer is a localized change with no effect on workflow engine behavior.

## Decisions

### Route all non-error run output through the Cobra writer

The validated workflow summary, schedule-not-implemented guidance, server-unavailable notice, engine-disabled notice, execution progress, completion status, and per-step output all render through `cmd.OutOrStdout()`.

### Replace stdout swapping in schedule-path tests

The existing schedule-path regression now captures the command writer directly instead of reassigning process-global stdout.

## Non-Goals

- No change to workflow run semantics
- No implementation of CLI schedule registration
