## Context

The repository has been steadily removing test dependencies on process-global stdin/stdout/stderr and routing user-facing output through explicit writers. `prompt.Passphrase(...)` and `ConfirmIO(...)` already fit that direction, but `prompt.Confirm(...)` still binds directly to `os.Stdin` and `os.Stdout`, which makes the wrapper itself harder to verify without global stream replacement.

The same cleanup slice also removed a shared CLI harness from `internal/testutil`. That was correct for the global stdout interception path, but several packages still rely on the package's lightweight config/bootstrap loader helpers for command construction.

## Goals / Non-Goals

**Goals:**
- Make `prompt.Confirm(...)` use injectable default streams in tests
- Preserve the existing runtime behavior for callers that rely on package defaults
- Keep `ConfirmIO(...)` unchanged for callers that already provide explicit streams
- Restore the non-stdio shared loader helpers that existing CLI regressions still depend on

**Non-Goals:**
- Changing prompt text or confirmation semantics
- Migrating callers away from `ConfirmIO(...)`
- Introducing new CLI flags or user-facing behavior

## Decisions

Use package-level `io.Reader` / `io.Writer` seams for the default confirmation wrapper.
Rationale: this matches the existing seam pattern already used in `internal/cli/prompt` for hidden passphrase input and keeps the wrapper change small.
Alternative considered: deleting `Confirm(...)` and forcing all callers to use `ConfirmIO(...)`. Rejected because it would create unnecessary call-site churn for a narrow testability issue.

Keep `ConfirmIO(...)` as the single implementation of the prompt/read/parse flow.
Rationale: the wrapper should only supply defaults. Reusing `ConfirmIO(...)` avoids duplicate parsing logic and keeps behavior identical across wrapper and explicit-stream call paths.

Reintroduce only the config/bootstrap loader helpers in `internal/testutil`, not the removed stdout interception harness.
Rationale: the loader helpers are still broadly useful for CLI tests and do not conflict with the newer command-writer-only test pattern.
Alternative considered: replacing every remaining `testutil.FakeCfgLoader` / `FakeBootLoader` call site in the same turn. Rejected because it is broader than needed to restore full-suite verification for this change.

## Risks / Trade-offs

- [Risk] Package-level seams can leak between tests if not restored. → Mitigation: restore original seam values with `t.Cleanup(...)` in every regression test.
- [Trade-off] The change adds a small amount of package-global test indirection. → Mitigation: keep the seams limited to the wrapper defaults and continue using `ConfirmIO(...)` for explicit call paths.
