## Context

The workflow run schedule tests cover several mutually exclusive execution paths. One of those tests temporarily replaces the package-global `executeWorkflowDirect` seam. Because the tests all use `t.Parallel()`, that replacement can bleed into sibling tests and flip their expected output.

## Goals / Non-Goals

**Goals:**
- Remove the race between workflow run schedule tests
- Preserve the existing assertions and coverage
- Restore deterministic `go test ./...` behavior

**Non-Goals:**
- Refactoring workflow runtime code
- Reworking the seam into a larger dependency-injection change
- Changing any user-facing workflow CLI behavior

## Decisions

Remove parallel execution from the workflow run schedule tests in the file that shares the mutable seam.
Rationale: it is the smallest change that fully eliminates the race without touching production code.

## Risks / Trade-offs

- [Trade-off] The affected tests lose some parallelism. → Mitigation: the file is small, and deterministic correctness is more valuable than marginal test speed here.
