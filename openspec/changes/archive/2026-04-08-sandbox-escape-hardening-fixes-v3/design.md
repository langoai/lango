## Context

The third Codex review pass on top of `sandbox-escape-hardening-fixes-v2` identified three issues in `internal/cli/sandbox/sandbox.go`. Two of them (graceful degradation and the backend column lie) are regressions introduced by my own work in earlier rounds. The third (touch portability) predates this PR but is small and lives in the same file, so it fits naturally with the other CLI fixes in this round.

This change is intentionally narrow: only one source file and its test file are touched.

## Goals / Non-Goals

**Goals:**
- `lango sandbox status` renders the configuration / capabilities / backend availability sections in every degraded mode where it previously errored out.
- The Recent Sandbox Decisions section is silently skipped when the audit DB is unreachable.
- `lango sandbox status` continues to run at most one bootstrap pass per invocation in the healthy path (preserving the round-1 v2 fix).
- The Recent Decisions row formatter never falsely advertises a sandbox backend for `excluded` / `skipped` / `rejected` verdicts.
- The smoke tests cannot produce a false PASS on hosts where `/usr/bin/touch` does not exist.

**Non-Goals (deferred to PR 5):**
- Subdirectory walk-up `.git` discovery.
- `MCPServerPolicy` workspace `.git` deny.
- `bwrap --version` probe false positives on hosts with `kernel.unprivileged_userns_clone=0`.
- Native Landlock+seccomp backend; file-level deny; symlink chain resolution; glob/path semantics normalization; per-tool policy overrides.

## Decisions

### D1 — `sandbox status` graceful degradation via cfgLoader fallback

**Problem**: Round 1 v2 collapsed `cfgLoader` and `bootLoader` into a single bootLoader to fix the double-passphrase regression where status called both loaders unconditionally and triggered `bootstrap.Run()` twice. The unintended consequence: the status command now errors out entirely whenever bootLoader is nil OR returns an error. Round 1's plan explicitly promised graceful degradation ("Recent Decisions silently skipped, the rest of status renders") and v1 broke that promise.

**Decision**: `newStatusCmd` accepts both `cfgLoader` and `bootLoader` again. The RunE tries `bootLoader` first. On success, `cfg = boot.Config` and the boot is remembered for the Recent Decisions query. On nil bootLoader OR a bootLoader error, the RunE falls back to `cfgLoader` to load the config independently. The Recent Decisions section is silently skipped in degraded modes. The "no double bootstrap" property of round 1 is preserved because the fallback only fires when the primary loader did NOT actually run a bootstrap to completion.

**Failure envelope**:
- Healthy: bootLoader succeeds → 1 bootstrap, full output, Recent Decisions visible.
- Nil bootLoader (test wiring): cfgLoader runs → 1 bootstrap, full output minus Recent Decisions.
- bootLoader errors: fallback to cfgLoader → may also fail (same `bootstrap.Run`), but if it succeeds, status renders the non-audit sections. If it also fails, status errors out (no worse than v1).
- Both loaders nil: explicit error.

**Test coverage**: `TestNewStatusCmd_NilBootLoaderFallsBackToCfgLoader` and `TestNewStatusCmd_BootLoaderErrorFallsBackToCfgLoader` use a stub `cfgLoader` returning `defaultTestConfig()` and verify that the rendered output contains the config / active isolation / backend availability headers but NOT the "Recent Sandbox Decisions" header.

### D2 — Backend column rule: `-` for non-applied verdicts

**Problem**: The Recent Decisions row formatter previously printed whatever `Backend` value was in the audit row's `Details["backend"]` map, falling back to `-` only when the value was empty. Every publish site (`exec`, `skill`, `mcp`) calls `publishSandboxDecision` which stamps `Backend` from the wired isolator's `Name()` regardless of which decision was made — so audit rows for `excluded` commands carry `backend=bwrap`, rows for `skipped` carry `backend=seatbelt`, etc. The user sees `excluded  bwrap  git status` and reasonably concludes the command was sandboxed under bwrap. It was NOT.

**Decision**: Force the backend column to `-` whenever `decision != "applied"` OR backend is empty. Only `applied` decisions actually ran inside a sandbox backend; everything else (excluded / skipped / rejected) ran unsandboxed (or did not run at all). The display must reflect that.

**Implementation**: Extracted `formatDecisionLine(ts, sessShort, decision, backend, target, reason) string` from the row rendering loop so the rule can be unit-tested directly without an ent.Client + audit fixture. The publish sites are intentionally NOT changed — keeping them stamp Backend uniformly is simpler than threading per-decision logic into every call site, and the display layer is the right place to apply the verdict-specific formatting.

**Test coverage**: `TestFormatDecisionLine_BackendColumnForNonAppliedRows` table-tests five cases (`applied`+backend → keep, `excluded`/`skipped`/`rejected`+backend → `-`, `applied`+empty → `-`). For non-applied verdicts the test additionally asserts that the original Backend value does NOT appear anywhere in the rendered line.

### D3 — `findTouch` portability for smoke tests

**Problem**: `runWriteTest` and `runWorkspaceWriteTest` called `exec.Command("/usr/bin/touch", ...)`. On non-merged-/usr Linux distributions (BusyBox, Alpine, embedded) the binary lives at `/bin/touch` instead, or only in `$PATH` via a busybox alias. Two failure modes:
1. `runWorkspaceWriteTest` always returns false because `c.Run()` returns ENOENT, even when the sandbox would have allowed the write. The user sees a FAIL with no actionable signal.
2. `runWriteTest` returns `c.Run() != nil` which is true on ENOENT — producing a **false PASS** for write-deny even when the sandbox is doing nothing. This is the worse failure: the smoke test reports green while the sandbox is silently broken.

**Decision**: Add a `findTouch()` helper that uses `exec.LookPath("touch")` first, then falls back to explicit candidates `/usr/bin/touch` and `/bin/touch`. Both call sites use the helper. On missing touch, both tests return `false` so neither produces a false positive. The runWriteTest false-PASS path is closed.

**Why not extend the test runner with a SKIP state**: Possible, but adds bool→tri-state plumbing through the test loop and label/passOK/failOK struct. The minimal fix (return false when touch is unfindable) closes the false-PASS without that plumbing. Future PR can add SKIP if needed.

**Test coverage**: `TestFindTouch` verifies the helper returns a non-empty path on every supported test environment (or skips if no touch is reachable, which is acceptable for the unit test even though the smoke test cannot skip).

## Risks / Trade-offs

- **D1 reintroduces both loaders to `sandbox status`** but is NOT a regression of the round-1 double-bootstrap fix. The status RunE tries `bootLoader` FIRST and only falls back to `cfgLoader` when bootLoader is nil or errors. In the healthy path only one bootstrap runs.
- **D2 trusts the display layer to enforce the backend rule** rather than the publish sites. If a future audit consumer reads the raw `Details["backend"]` field directly, it will still see the published value (`bwrap`/`seatbelt`/etc) for non-applied verdicts. Mitigation: the rule is documented in the spec requirement and in the formatter doc comment.
- **D3 returns `false` for missing touch** rather than SKIP. On hosts without touch, the smoke test will print FAIL — better than the previous false PASS but still misleading. A future PR can add SKIP plumbing if needed.

## Migration Plan

No schema migration, no config migration. The fix is purely code + tests. After deployment:
- `lango sandbox status` works in degraded modes that previously errored out.
- Existing audit rows for excluded / skipped / rejected commands re-render with `-` in the backend column on the next status invocation.
- `lango sandbox test` produces accurate results on BusyBox / Alpine.

Rollback: revert the commit. No data to undo.
