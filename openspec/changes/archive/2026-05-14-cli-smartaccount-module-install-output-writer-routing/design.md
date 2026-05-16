## Overview

`lango account module install` depends on a live smart account bootstrap and on-chain manager call. A narrow seam around the final installation result is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small module-install seam

The seam accepts the parsed module type and address, performs the real installation in production, and returns the transaction hash plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route success output through the Cobra writer

- Success text uses `cmd.OutOrStdout()`

## Non-Goals

- No change to module installation semantics
- No change to module installation output content
