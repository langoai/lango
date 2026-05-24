## Overview

The chat entrypoint startup path mirrors the same boot/logging/banner pattern as other TUI commands. A small stderr seam plus injected builders is enough to make that path deterministic under test.

## Decision

- Keep the visible startup text unchanged
- Replace direct `os.Stderr` writes with the chat stderr seam
- Reuse the same seam pattern already applied to cockpit and workbench

## Consequences

- All interactive top-level TUI entrypoints now share the same startup-notice capture model
- Future regressions in the chat startup notice path fail fast in `cmd/lango` tests
