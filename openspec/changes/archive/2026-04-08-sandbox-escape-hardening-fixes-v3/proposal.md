## Why

A third Codex review pass on `sandbox-escape-hardening-fixes-v2` (round 2) found three more issues, all in `internal/cli/sandbox/sandbox.go`. None block compilation, but each makes the new diagnostics or smoke tests behave incorrectly in real environments.

1. **P2 — `sandbox status` aborts when bootLoader fails.** Round 1 collapsed cfgLoader and bootLoader into a single bootLoader to fix the double-passphrase regression. The unintended consequence: nil bootLoader OR a bootLoader error (signed-out, locked DB, non-interactive) errors out the entire status command, hiding the configuration / capabilities / backend availability sections that do not depend on the audit DB. Round 1's plan promised graceful degradation; v1 broke it.
2. **P3 — Recent Decisions backend column lies.** The row formatter prints whatever `Backend` value is in the audit row's details map, even for `excluded` / `skipped` / `rejected` decisions. Every publish site stamps `Backend` from the wired isolator regardless of decision, so audit rows for excluded commands carry `backend=bwrap`. Result: rows like `excluded  bwrap  git status` falsely suggest the command was sandboxed under bwrap, when in fact it ran unsandboxed.
3. **P2 — `sandbox test` smoke tests hardcode `/usr/bin/touch`.** `runWriteTest` and `runWorkspaceWriteTest` call `exec.Command("/usr/bin/touch", ...)` directly. On non-merged-/usr layouts (BusyBox, Alpine) the binary is absent, so `runWorkspaceWriteTest` always FAILs and `runWriteTest` returns `c.Run() != nil` which is true on ENOENT — producing a **false PASS** for the write-deny case even when the sandbox is doing nothing.

## What Changes

- **`internal/cli/sandbox/sandbox.go`**:
  - `newStatusCmd` accepts `(cfgLoader, bootLoader)` again. It tries `bootLoader` first so a single bootstrap pass serves both config rendering and the Recent Decisions audit query (preserves the round-1 fix to the double-passphrase regression). On nil bootLoader OR a bootLoader error, falls back to `cfgLoader` so the config / capabilities / backend availability sections still render. Recent Decisions section is silently skipped in degraded modes.
  - Extracted `formatDecisionLine` from the row rendering loop. Forces the backend column to `-` whenever `decision != "applied"` (or backend is empty) so excluded / skipped / rejected rows do not falsely advertise a sandbox backend that did not actually run the command.
  - New `findTouch()` helper that locates `touch` via `exec.LookPath` with explicit fallbacks to `/usr/bin/touch` and `/bin/touch`. `runWriteTest` and `runWorkspaceWriteTest` use it; both return false on missing touch so the smoke tests cannot produce a false PASS on ENOENT.
- **`internal/cli/sandbox/sandbox_test.go`**: New `TestFormatDecisionLine_BackendColumnForNonAppliedRows` (5 cases), `TestNewStatusCmd_NilBootLoaderFallsBackToCfgLoader`, `TestNewStatusCmd_BootLoaderErrorFallsBackToCfgLoader`, `TestFindTouch`. Helper additions: `defaultTestConfig()`, `errSimulatedBootFailure`.

## Capabilities

### New Capabilities

(none — pure fix change)

### Modified Capabilities

- `os-sandbox-core`: Add two new requirements — `Sandbox status graceful degradation` (cfgLoader fallback contract) and `Sandbox decision row formatter` (backend column rules).

## Impact

- **Affected code**: `internal/cli/sandbox/sandbox.go`, `internal/cli/sandbox/sandbox_test.go`.
- **Affected specs**: `os-sandbox-core` (two new requirements added).
- **Documentation**: No README / docs / prompts changes.
- **Runtime behavior**:
  - `lango sandbox status` continues to render config / capabilities / backend availability sections in degraded modes (signed-out, locked DB, no bootLoader). Recent Decisions section is silently skipped instead of aborting the whole command.
  - Recent Decisions rows show `-` in the backend column for excluded / skipped / rejected verdicts so users cannot mistake an excluded command for a sandboxed one.
  - `lango sandbox test` works on BusyBox / Alpine / non-merged-/usr environments via `exec.LookPath("touch")` with explicit fallbacks; cannot produce a false PASS on ENOENT.
- **Out of scope (PR 5)**: subdirectory walk-up `.git` discovery; MCP workspace `.git` deny; `bwrap --version` probe false positive.
