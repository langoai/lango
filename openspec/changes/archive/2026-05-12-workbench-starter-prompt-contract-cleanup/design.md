## Overview

This is a behavior-preserving architecture cleanup for the workbench quick-start path.

## Design Decisions

### One helper owns prompt generation

The new `internal/cli/workbenchstart` package owns:

- default prompt fallbacks
- workspace-aware prompt generation
- Git-aware prompt sharpening

That keeps prompt generation independent from both the workbench shell and the Mission Control page rendering logic.

### The page consumes, it does not invent

The Mission Control page still accepts an explicit prompt set from workbench wiring, but its local fallback now uses the same shared helper defaults. This removes duplicate product copy constants without changing runtime behavior.
