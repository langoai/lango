## Overview

This change generalizes the recent CLI harness cleanup into a repository-wide invariant for test files.

## Decision

- Scan `_test.go` files under `cmd/` and `internal/`
- Reject process-global stdio reassignment
- Reject legacy `testutil.ExecCmd(...)` and `testutil.ExecCmdOK(...)`
- Exclude the guard files themselves so the detector can describe the forbidden strings without self-failing

## Consequences

- Harness regressions fail fast outside CLI too
- Reviewers do not need to re-enforce the same low-signal rule manually across the test suite
