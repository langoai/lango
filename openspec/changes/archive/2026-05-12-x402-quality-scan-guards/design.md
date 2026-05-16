## Overview

This change converts an informal static-code expectation into a direct test gate.

## Design Decisions

### File scanning instead of shelling out

The guard reads Go source files directly from the repository tree. That avoids runtime dependencies on `rg`, `grep`, or other host tools while still making the requirement executable in CI.

### Self-trigger resistance

The test constructs the forbidden token strings indirectly so the guard does not fail on its own source code.

### Scope follows the spec

The `context.TODO()` guard is limited to the `internal/x402` package source, while the legacy client-factory reference guard scans the whole repository to catch both implementation and call-site regressions.
