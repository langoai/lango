## Overview

Use Cobra command streams for all operator-visible `security secrets` I/O.

## Decisions

- use a temp file-backed bootloader in tests so `set` then `list`/`delete` can observe persistent state
- `set --value-hex` remains the non-interactive test path
- `delete` prompt uses `cmd.OutOrStdout()` and `cmd.InOrStdin()`
- `list` and `set` success output use `cmd.OutOrStdout()`

## Risks

- none beyond test setup complexity, mitigated by a local bootloader helper in tests
