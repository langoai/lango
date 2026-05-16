## Overview

The passphrase acquisition path has two non-interactive branches that are worth isolating for tests: stdin-pipe reading and keyring warning emission. A small internal helper plus a reader-based stdin helper is enough to cover both without changing public behavior.

## Decisions

### Add `ReadStdinPipeFromReader(...)`

The existing `ReadStdinPipe()` remains as the production entry point, but now delegates to a reader-based helper that tests can call directly.

### Add an internal `acquireWithIO(...)` helper

`Acquire(...)` now delegates to an internal helper that accepts injected stdin, stderr, and the terminal/non-terminal decision. Tests can exercise the non-interactive branches through that helper without replacing global process streams.

## Non-Goals

- No change to the user-facing passphrase acquisition priority chain
- No change to interactive hidden-input handling
