## Overview

The warning path lives entirely inside `readDBStatusNonInteractive(...)`, so small package-level seams are sufficient. The command's user-visible behavior remains the same; only the destination writer and acquisition function become injectable under test.

## Decisions

### Add package-level seams

- `acquireNonInteractivePassphrase`
- `statusErrWriter`

These default to the current production implementations.

## Non-Goals

- No change to security status semantics
- No change to graceful-degrade behavior when DB-backed fields are unavailable
