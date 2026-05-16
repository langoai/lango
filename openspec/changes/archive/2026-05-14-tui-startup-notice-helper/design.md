## Overview

The startup-notice path is already seam-aware and regression-tested. This change just removes repeated rendering code by extracting the shared three-line pattern into a helper.

## Decision

- Keep the exact banner/log/initializing output unchanged
- Use a single helper for chat, cockpit, and workbench
- Rely on the existing command-level regressions as proof of behavior preservation

## Consequences

- Lower maintenance cost for future startup-notice edits
- Less chance of one TUI entrypoint drifting from the others
