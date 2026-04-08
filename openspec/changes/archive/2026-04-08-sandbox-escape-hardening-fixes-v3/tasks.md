## 1. sandbox status graceful degradation (Stage 1)
- [x] 1.1 `newStatusCmd` accepts `(cfgLoader, bootLoader)` again. Try `bootLoader` first; on nil OR error, fall back to `cfgLoader` so config / capabilities / backend availability sections still render. Recent Decisions silently skipped in degraded mode.
- [x] 1.2 `NewSandboxCmd` re-wires `newStatusCmd(cfgLoader, bootLoader)` (external signature unchanged).
- [x] 1.3 Add `defaultTestConfig()` and `errSimulatedBootFailure` test helpers to `sandbox_test.go`.
- [x] 1.4 Add `TestNewStatusCmd_NilBootLoaderFallsBackToCfgLoader` verifying that the config / active isolation / backend availability sections render when bootLoader is nil and "Recent Sandbox Decisions" is silently skipped.
- [x] 1.5 Add `TestNewStatusCmd_BootLoaderErrorFallsBackToCfgLoader` verifying the same behaviour when bootLoader returns an error.

## 2. Recent Decisions backend column rule (Stage 2)
- [x] 2.1 Extract `formatDecisionLine(ts, sessShort, decision, backend, target, reason) string` from `renderRecentDecisions`'s row loop.
- [x] 2.2 Force `backend = "-"` whenever `decision != "applied"` OR backend is empty inside `formatDecisionLine`.
- [x] 2.3 Update `renderRecentDecisions` to call the new helper.
- [x] 2.4 Add `TestFormatDecisionLine_BackendColumnForNonAppliedRows` (5 cases) asserting both the backend column value and that the original Backend value does NOT leak into the rendered line for non-applied verdicts.

## 3. findTouch portability for smoke tests (Stage 3)
- [x] 3.1 Add `findTouch()` helper using `exec.LookPath("touch")` with `/usr/bin/touch` and `/bin/touch` fallbacks. Returns empty string when unreachable.
- [x] 3.2 `runWriteTest` calls `findTouch()`; returns false when empty so the false-PASS on ENOENT is closed.
- [x] 3.3 `runWorkspaceWriteTest` calls `findTouch()`; returns false when empty.
- [x] 3.4 Add `TestFindTouch` verifying the helper returns a non-empty path on the test host (skips if neither PATH nor fallbacks have touch).

## 4. OpenSpec change + sync + archive (Stage 4)
- [x] 4.1 `openspec new change sandbox-escape-hardening-fixes-v3`
- [x] 4.2 Write `proposal.md` covering round 3
- [x] 4.3 Write `design.md` with Decisions D1-D3 / Risks / Migration Plan / PR 5 deferred items
- [x] 4.4 Write delta spec `specs/os-sandbox-core/spec.md` with `## ADDED Requirements` for `Sandbox status graceful degradation` and `Sandbox decision row formatter`
- [x] 4.5 Write `tasks.md` (this file)
- [x] 4.6 `openspec validate sandbox-escape-hardening-fixes-v3 --strict`
- [ ] 4.7 `openspec archive sandbox-escape-hardening-fixes-v3 -y --no-validate` (project-wide main specs lack `## Purpose` headers, separate meta-fix issue)
- [ ] 4.8 Final verify: build cross-platform, test, lint
