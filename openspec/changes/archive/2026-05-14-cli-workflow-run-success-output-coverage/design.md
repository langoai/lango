## Overview

The workflow-run command now has good coverage for schedule and fallback guidance paths, but not for successful direct execution. A narrow execution seam is enough to stub the direct engine result and verify completion output through the command writer.

## Decisions

### Add a direct-execution seam

The seam accepts the boot result and parsed workflow, and returns a workflow run result or an error. Production still delegates to the real workflow engine through this seam.

### Verify the success-path output contract explicitly

Tests assert that the execution banner, completion status, and per-step output all appear on the command writer when the seam returns a successful result.

## Non-Goals

- No change to workflow engine semantics
- No implementation of additional execution retry/fallback behavior
