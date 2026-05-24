## Overview

The warning path is isolated to `warnFallbackOnce(...)`, so a single writer seam is enough to make the behavior deterministic in tests without affecting command execution or sandbox decisions.

## Decisions

### Add a package-level warning writer seam

`execWarningWriter` defaults to `os.Stderr` in production and can be redirected in tests.

### Test the one-shot contract directly

The regression exercises two warning attempts and verifies that only the first reason is emitted.

## Non-Goals

- No change to fail-open vs fail-closed policy semantics
- No change to event bus publication behavior
