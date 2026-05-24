## Overview

The workflow management commands are mostly read-only views over workflow state plus a single cancellation action. Writer-routing can be fixed in place for the read surfaces, while the cancel path benefits from a narrow execution seam so tests do not depend on a live workflow engine.

## Decisions

### Route read surfaces directly through the Cobra writer

List, status, and history now render tables, detail views, and empty-state messages through `cmd.OutOrStdout()`.

### Introduce a small workflow-cancel seam

The cancel seam wraps bootstrap and engine cancellation, then returns the final confirmation message. Tests can stub it without constructing a live engine.

## Non-Goals

- No change to workflow execution semantics
- No change to workflow run command output in this change
