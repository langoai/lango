## Overview

The non-interactive acquisition path is intentionally narrow, so the smallest useful seam is a helper that accepts the stderr writer. Production still delegates to it with `os.Stderr`, while tests can capture the warning behavior directly.

## Decisions

### Add `acquireNonInteractiveWithIO(...)`

`AcquireNonInteractive(...)` remains the public API and simply delegates to the internal helper with `os.Stderr`.

### Cover both warning and silent-fallback branches

Tests verify:
- non-`ErrNotFound` keyring errors emit a warning and still fall back
- `ErrNotFound` remains silent and still falls back

## Non-Goals

- No change to non-interactive source priority
- No change to caller-facing `ErrNoNonInteractiveSource` behavior
