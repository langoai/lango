## Overview

The memory CLI test suite already defines helpers for command writer capture, including input-aware variants. The remaining `testutil.ExecCmd` uses were limited to error-path assertions and can be replaced directly.

## Decisions

### Standardize on package-local command helpers

`executeMemoryCmd(...)` and `executeMemoryCmdWithInput(...)` now cover both success and error paths in the memory CLI tests.

## Non-Goals

- No runtime behavior changes
- No migration of unrelated packages in this change
