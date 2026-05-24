## Overview

This change turns the recent CLI test-harness cleanup into a repository invariant.

## Decision

- Scan only CLI `_test.go` files under `internal/cli`
- Reject direct reassignment of `os.Stdin`, `os.Stdout`, and `os.Stderr`
- Reject legacy `testutil.ExecCmd(...)` and `testutil.ExecCmdOK(...)` references

## Consequences

- Future CLI test regressions fail fast
- Command-stream testing remains the default path instead of drifting back to process-global interception
