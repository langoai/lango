## Why

PR 4 (`sandbox-escape-hardening`) was archived on 2026-04-08. A Codex code review run immediately after archive identified four regressions / gaps that cause PR 4's contract to NOT hold in practice:

1. **P1 — Seatbelt control-plane mask is a no-op on macOS.** PR 4 added `~/.lango` and `.git` to `DenyPaths`, but `seatbelt_profile.go` only emits `(deny file-write* ...)`. Combined with `(allow file-read*)` from `ReadOnlyGlobal=true`, sandboxed children still read `~/.lango/lango.db` (session DB + audit log), `.git/config`, and encrypted config tokens.
2. **P1 — Baseline `.git` deny breaks Linux bwrap for every non-repo workspace.** `DefaultToolPolicy` pushes `filepath.Join(workDir, ".git")` unconditionally, but `compileBwrapArgs` stat+IsDir-rejects missing or non-directory deny paths. Non-repo workspaces and linked worktrees (where `.git` is a file) make the entire sandbox apply fail — either rejected (failClosed=true) or silently fallen-back (failClosed=false).
3. **P2 — `lango sandbox status` triggers two independent `bootstrap.Run` calls per invocation**, prompting for the passphrase twice.
4. **P2 — `skill.Executor.executeScript` early-returns on `isolator==nil && failClosed`**, skipping the `SandboxDecisionEvent{Decision:"rejected"}` publish that PR 4's `sandbox-exception-policy` spec explicitly requires.

This fix PR makes PR 4 actually work as documented. It does not add new features.

## What Changes

- **`internal/sandbox/os/seatbelt_profile.go`**: Template emits both `(deny file-read* (subpath "..."))` and `(deny file-write* (subpath "..."))` for every `DenyPaths` entry so the control-plane mask blocks reads (not just writes).
- **`internal/sandbox/os/policy.go`**: Add `isDir` private helper. `DefaultToolPolicy` and `MCPServerPolicy` now gate baseline deny entries (`.git`, `dataRoot`) on `isDir(path)` so missing or non-directory entries are silently skipped instead of propagating a failure through `compileBwrapArgs`.
- **`internal/skill/executor.go`**: Delete the early return at `isolator==nil && failClosed`. Consolidate the nil-isolator decision (both `rejected` and `skipped` branches) into a single publish path before temp file creation. Remove the now-dead else-if branch after `isolator.Apply`.
- **`internal/cli/sandbox/sandbox.go`**: `newStatusCmd` takes only the `BootLoader` (cfgLoader removed), derives cfg from `boot.Config`, and `defer boot.DBClient.Close()`s. `renderRecentDecisions` takes the pre-resolved `*bootstrap.Result` directly instead of re-invoking the loader.
- New regression tests covering: Seatbelt read+write deny, missing `.git`, `.git` as a file (worktree), missing dataRoot, non-repo bwrap compile, worktree bwrap compile, skill fail-closed rejection publish, skill fail-open skipped publish.

## Capabilities

### New Capabilities

(none — this is a pure fix PR)

### Modified Capabilities

- `os-sandbox-core`: Modify the `Default tool policy` baseline deny scenario so `.git` is only denied when it exists as a directory; add scenarios for missing `.git` and worktree `.git` file; add a scenario asserting that Seatbelt profile generation emits both `file-read*` and `file-write*` deny rules for every `DenyPaths` entry.

## Impact

- **Affected code**: `internal/sandbox/os/seatbelt_profile.go`, `internal/sandbox/os/policy.go`, `internal/sandbox/os/policy_test.go`, `internal/sandbox/os/bwrap_args_test.go`, `internal/skill/executor.go`, `internal/skill/executor_test.go`, `internal/cli/sandbox/sandbox.go`, `internal/cli/sandbox/sandbox_test.go`.
- **Affected specs**: `os-sandbox-core` (one existing scenario modified, two scenarios added).
- **Documentation**: No README / docs / prompts changes required. PR 4 docs already describe the correct contract; this PR makes the code match.
- **Runtime behavior**:
  - macOS: sandboxed exec/skill/MCP children can no longer read `~/.lango` or `.git` contents.
  - Linux: `compileBwrapArgs` no longer fails for non-repo workspaces or worktree checkouts.
  - `lango sandbox status` prompts for the passphrase once per invocation.
  - skill `rejected` decisions are persisted in audit.
- **Out of scope (PR 5)**: `bwrap --version` probe false positive on hosts with unprivileged userns disabled; native Landlock+seccomp backend; file-level deny via `--ro-bind /dev/null <file>`; symlink chain resolution; glob/path semantics normalization; per-tool policy override.
