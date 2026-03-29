# Design: Restore dev code lost during worktree file copy

## Root Cause

Worktree workers (`isolation: "worktree"`) create git worktrees based on `origin/main`, not the current feature branch. When worker output files were copied (`cp`) to the feature branch, they overwrote files that contained `dev`-only code not present in `main`.

## Restoration Strategy

For each affected file: restore `dev` branch version as base, then re-apply intentional additions on top.

### D1: loader.go — LoadResult + Context defaults
`dev` has `LoadResult` struct, `Load()` returning `(*LoadResult, error)`, context profile/auto-enable logic, Retrieval defaults, and contextProfile validation. Working tree had reverted to `main`'s simplified `Load()` returning `(*Config, error)`.

**Restore:** Full `dev` version. **Add:** `Context: ContextConfig{Allocation: ...}` block with spec defaults (0.30/0.25/0.25/0.10/0.10).

### D2: state_update.go — 35 + 22 field handlers
`dev` has 308 case handlers. Working tree had 291 (17 missing: Orchestration 8, RunLedger 9, Provenance 5, OS Sandbox 9, TraceStore 4 — but some overlap makes 35 total missing).

**Restore:** Full `dev` version. **Add:** 22 new handlers for Context Profile, Retrieval, Auto-Adjust, Context Budget.

### D3: setup_flow.go — 3 + 4 form factory cases
`dev` has `runledger`, `provenance`, `os_sandbox` cases. Working tree had replaced them with only the 4 new cases.

**Restore:** Full `dev` version. **Add:** 4 new cases (context_profile, retrieval, auto_adjust, context_budget) after existing `librarian` case.

### D4: migrate.go — LoadResult handling
`dev` uses `result.Config` and `result.ExplicitKeys` from `config.Load()`. Working tree had `cfg, nil` because `Load()` was incorrectly returning `*Config`.

**Restore:** Full `dev` version. No additions needed.

## Files Changed

| File | Restore from dev | Additions |
|------|-----------------|-----------|
| `internal/config/loader.go` | LoadResult, Load(), profile logic, Retrieval defaults, validation | Context allocation defaults |
| `internal/cli/tuicore/state_update.go` | 35 field handlers | 22 new handlers |
| `internal/cli/settings/setup_flow.go` | 3 form factory cases | 4 new cases |
| `internal/configstore/migrate.go` | LoadResult handling | None |
