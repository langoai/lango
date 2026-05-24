## Overview

This is a behavior-preserving architecture simplification for the workbench quick-start path.

## Design Decisions

### Only stable inputs cross the dependency boundary

`workDir` is a stable environment input. The prompt slice is derived data. The page now receives only the stable input and uses the shared helper to derive the prompt set as needed.

### Shared helper remains the only source of truth

The `internal/cli/workbenchstart` package still owns all prompt generation logic. This change only removes duplicate transport state, not the shared generation contract itself.
