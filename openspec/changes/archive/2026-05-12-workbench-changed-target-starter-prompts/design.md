## Overview

This change makes the dirty-repository quick-start prompt point closer to the operator's actual edit surface.

## Design Decisions

### Changed-target extraction stays lightweight

The workbench reuses the existing `git status --short` call and parses only enough information to identify up to three distinct changed targets. It does not invoke diff, blame, or history analysis.

### Summaries use top-level targets

When possible, changed paths are normalized to the top-level file or directory so the prompt stays short and legible in the TUI. This favors clarity over exhaustiveness.

### Dirty-state fallback remains intact

If changed targets cannot be parsed, the prompt still falls back to the existing branch-and-dirty wording. Startup guidance improves when the signal is available but never depends on it.
