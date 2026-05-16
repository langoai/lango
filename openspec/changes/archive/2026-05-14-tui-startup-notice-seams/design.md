## Overview

The startup notice path for the TUI entrypoints is short and deterministic: resolve boot, init logging, print banner/log path/initializing text, then build the app. Extracting just enough seams around that path lets tests assert the notices without needing to run Bubble Tea or the full app runtime.

## Decision

- Keep the human-visible startup notice text unchanged
- Replace direct `os.Stderr` writes with injected stderr writers for cockpit and workbench
- Introduce boot/logging/app-builder seams only where needed for focused startup regressions

## Consequences

- Wrapper-driven captures can exercise startup notices deterministically
- Future refactors that break startup notice routing will fail fast in `cmd/lango` tests
